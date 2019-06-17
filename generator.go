package main

import (
	"bytes"
	"go/format"
	"html/template"
	"log"
	"path"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

type targetFile struct {
	Name     string
	Pkg      string
	Services []*targetService
}

type targetService struct {
	Name    string
	Methods []*targetMethod
}

type targetMethod struct {
	Name string
	Arg  string
}

func Generate(req *plugin.CodeGeneratorRequest) (*plugin.CodeGeneratorResponse, error) {
	// Filter files to target files.
	targetFiles := make([]*targetFile, 0)
	for _, f := range req.GetProtoFile() {
		target := genTarget(f)
		if target != nil {
			targetFiles = append(targetFiles, target)
		}
	}

	// Generate response files from proto files.
	respFiles := make([]*plugin.CodeGeneratorResponse_File, 0)
	for _, f := range targetFiles {
		respFiles = append(respFiles, genRespFile(f))
	}

	// Format response files content.
	for _, f := range respFiles {
		content, err := format.Source([]byte(f.GetContent()))
		if err != nil {
			return nil, err
		}
		f.Content = proto.String(string(content))
	}

	return &plugin.CodeGeneratorResponse{
		File: respFiles,
	}, nil
}

func genTarget(file *descriptor.FileDescriptorProto) *targetFile {
	if len(file.GetService()) == 0 {
		return nil
	}
	f := &targetFile{
		Name:     file.GetName(),
		Pkg:      file.GetOptions().GetGoPackage(),
		Services: make([]*targetService, 0),
	}

	for _, service := range file.GetService() {
		s := &targetService{
			Name:    service.GetName(),
			Methods: make([]*targetMethod, 0),
		}
		for _, method := range service.GetMethod() {
			// Not generate streaming method
			if method.GetServerStreaming() || method.GetClientStreaming() {
				continue
			}
			s.Methods = append(s.Methods, &targetMethod{
				Name: method.GetName(),
				Arg:  ioname(method.GetInputType()),
			})
		}
		// Add if Service has a method
		if len(s.Methods) <= 0 {
			continue
		}
		f.Services = append(f.Services, s)
	}

	// Generate if File has a service
	if len(f.Services) <= 0 {
		return nil
	}
	return f
}

func genRespFile(target *targetFile) *plugin.CodeGeneratorResponse_File {
	buf := &bytes.Buffer{}

	t := template.Must(template.New("gohttp").Parse(codeTemplate))
	if err := t.Execute(buf, target); err != nil {
		log.Println("executing template:", err)
		panic(err)
	}

	return &plugin.CodeGeneratorResponse_File{
		Name:    proto.String(basename(target.Name) + ".http.go"),
		Content: proto.String(buf.String()),
	}
}

func basename(name string) string {
	if ext := path.Ext(name); ext == ".proto" || ext == ".protodevel" {
		name = name[:len(name)-len(ext)]
	}
	return name
}

func ioname(name string) string {
	s := strings.Split(name, ".")
	return s[len(s)-1]
}

const codeTemplate = `package {{ .Pkg }}

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)
{{ range $i, $service := .Services }}
type {{ $service.Name }}HTTPConverter struct {
	srv {{ $service.Name }}Server
}

func New{{ $service.Name }}HTTPConverter(srv {{ $service.Name }}Server) *{{ $service.Name }}HTTPConverter {
	return &{{ $service.Name }}HTTPConverter{
		srv: srv,
	}
}
{{ range $j, $method := $service.Methods }}
func (h *{{ $service.Name }}HTTPConverter) {{ $method.Name }}(cb func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error)) http.HandlerFunc {
	if cb == nil {
		cb = func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error) {
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				p := status.New(codes.Unknown, err.Error()).Proto()
				switch r.Header.Get("Content-Type") {
				case "application/protobuf", "application/x-protobuf":
					buf, err := proto.Marshal(p)
					if err != nil {
						return
					}
					if _, err := io.Copy(w, bytes.NewBuffer(buf)); err != nil {
						return
					}
				case "application/json":
					if err := json.NewEncoder(w).Encode(p); err != nil {
						return
					}
				default:
				}
			}
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		arg := &{{ $method.Arg }}{}
		contentType := r.Header.Get("Content-Type")
		if r.Method != http.MethodGet {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				cb(ctx, w, r, nil, nil, err)
				return
			}

			switch contentType {
			case "application/protobuf", "application/x-protobuf":
				if err := proto.Unmarshal(body, arg); err != nil {
					cb(ctx, w, r, nil, nil, err)
					return
				}
			case "application/json":
				if err := jsonpb.Unmarshal(bytes.NewBuffer(body), arg); err != nil {
					cb(ctx, w, r, nil, nil, err)
					return
				}
			default:
				w.WriteHeader(http.StatusUnsupportedMediaType)
				_, err := fmt.Fprintf(w, "Unsupported Content-Type: %s", contentType)
				cb(ctx, w, r, nil, nil, err)
				return
			}
		}

		ret, err := h.srv.{{ $method.Name }}(ctx, arg)
		if err != nil {
			cb(ctx, w, r, arg, nil, err)
			return
		}

		accept := r.Header.Get("Accept")
		if accept == "*/*" || accept == ""{
			if contentType != "" {
				accept = contentType
			} else {
				accept = "application/json"
			}
		}

		switch accept {
		case "application/protobuf", "application/x-protobuf":
			buf, err := proto.Marshal(ret)
			if err != nil {
				cb(ctx, w, r, arg, ret, err)
				return
			}
			if _, err := io.Copy(w, bytes.NewBuffer(buf)); err != nil {
				cb(ctx, w, r, arg, ret, err)
				return
			}
		case "application/json":
			m := jsonpb.Marshaler{
				EnumsAsInts:  true,
				EmitDefaults: true,
			}
			if err := m.Marshal(w, ret); err != nil {
				cb(ctx, w, r, arg, ret, err)
				return
			}
		default:
			w.WriteHeader(http.StatusUnsupportedMediaType)
			_, err := fmt.Fprintf(w, "Unsupported Accept: %s", accept)
			cb(ctx, w, r, arg, ret, err)
			return
		}
		cb(ctx, w, r, arg, ret, nil)
	})
}

func (h *{{ $service.Name }}HTTPConverter) {{ $method.Name }}WithName(cb func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error)) (string, string, http.HandlerFunc) {
	return "{{ $service.Name }}", "{{ $method.Name }}", h.{{ $method.Name }}(cb)
}{{ end }}{{ end }}
`
