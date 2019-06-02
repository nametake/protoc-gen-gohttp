package main

import (
	"bytes"
	"fmt"
	"go/format"
	"html/template"
	"io"
	"log"
	"path"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

type Generator struct{}

func New() *Generator {
	return &Generator{}
}

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

func p(w io.Writer, format string, args ...interface{}) {
	if w == nil {
		return
	}
	if _, err := fmt.Fprintf(w, format, args...); err != nil {
		panic(err)
	}
	if _, err := fmt.Fprintln(w); err != nil {
		panic(err)
	}
}

func (g *Generator) Generate(req *plugin.CodeGeneratorRequest) (*plugin.CodeGeneratorResponse, error) {
	// Filter files to target files.
	targetFiles := make([]*targetFile, 0)
	for _, f := range req.GetProtoFile() {
		target := g.genTarget(f)
		if target != nil {
			targetFiles = append(targetFiles, target)
		}
	}

	// Generate response files from proto files.
	respFiles := make([]*plugin.CodeGeneratorResponse_File, 0)
	for _, f := range targetFiles {
		respFiles = append(respFiles, g.genRespFile(f))
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

func (g *Generator) genTarget(file *descriptor.FileDescriptorProto) *targetFile {
	if len(file.GetService()) == 0 {
		return nil
	}
	tFile := &targetFile{
		Name:     file.GetName(),
		Pkg:      file.GetOptions().GetGoPackage(),
		Services: make([]*targetService, 0),
	}

	for _, s := range file.GetService() {
		service := &targetService{
			Name:    s.GetName(),
			Methods: make([]*targetMethod, 0),
		}
		for _, m := range s.GetMethod() {
			method := &targetMethod{
				Name: m.GetName(),
				Arg:  m.GetInputType(),
			}
			if !m.GetServerStreaming() && !m.GetClientStreaming() {
				service.Methods = append(service.Methods, method)
			}
		}
		if len(service.Methods) > 0 {
			tFile.Services = append(tFile.Services, service)
		}
		tFile.Services = append(tFile.Services, service)
	}

	if len(tFile.Services) <= 0 {
		return nil
	}

	return tFile
}

func (g *Generator) genRespFile(target *targetFile) *plugin.CodeGeneratorResponse_File {
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

func (g *Generator) writePackage(w io.Writer, f *descriptor.FileDescriptorProto) {
	p(w, "package %s", f.Options.GetGoPackage())
	p(w, "")
}

func (g *Generator) writeImports(w io.Writer) {
	p(w, "import (")
	p(w, "	\"bytes\"")
	p(w, "	\"context\"")
	p(w, "	\"encoding/json\"")
	p(w, "	\"fmt\"")
	p(w, "	\"io\"")
	p(w, "	\"io/ioutil\"")
	p(w, "	\"net/http\"")
	p(w, "")
	p(w, "	\"github.com/golang/protobuf/jsonpb\"")
	p(w, "	\"github.com/golang/protobuf/proto\"")
	p(w, "	\"google.golang.org/grpc/codes\"")
	p(w, "	\"google.golang.org/grpc/status\"")
	p(w, ")")
	p(w, "")
}

func (g *Generator) writeService(w io.Writer, s *descriptor.ServiceDescriptorProto) {
	p(w, "type %sHTTPConverter struct{", s.GetName())
	p(w, "	srv %sServer", s.GetName())
	p(w, "}")
	p(w, "")
	p(w, "func New%sHTTPConverter(srv %sServer) *%sHTTPConverter {", s.GetName(), s.GetName(), s.GetName())
	p(w, "	return &%sHTTPConverter{", s.GetName())
	p(w, "		srv: srv,")
	p(w, "	}")
	p(w, "}")
	p(w, "")
}

func (g *Generator) writeMethod(w io.Writer, s *descriptor.ServiceDescriptorProto, m *descriptor.MethodDescriptorProto) {
	p(w, "func (h *%sHTTPConverter) %s(cb func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error)) http.HandlerFunc {", s.GetName(), m.GetName())
	p(w, "	if cb == nil {")
	p(w, "		cb = func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error) {")
	p(w, "			if err != nil {")
	p(w, "				w.WriteHeader(http.StatusInternalServerError)")
	p(w, "				p := status.New(codes.Unknown, err.Error()).Proto()")
	p(w, "				switch r.Header.Get(\"Content-Type\") {")
	p(w, "				case \"application/protobuf\", \"application/x-protobuf\":")
	p(w, "					buf, err := proto.Marshal(p)")
	p(w, "					if err != nil {")
	p(w, "						return")
	p(w, "					}")
	p(w, "					if _, err := io.Copy(w, bytes.NewBuffer(buf)); err != nil {")
	p(w, "						return")
	p(w, "					}")
	p(w, "				case \"application/json\":")
	p(w, "					if err := json.NewEncoder(w).Encode(p); err != nil {")
	p(w, "						return")
	p(w, "					}")
	p(w, "				default:")
	p(w, "				}")
	p(w, "			}")
	p(w, "		}")
	p(w, "	}")
	p(w, "	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {")
	p(w, "		ctx := r.Context()")
	p(w, "")
	p(w, "		body, err := ioutil.ReadAll(r.Body)")
	p(w, "		if err != nil {")
	p(w, "			cb(ctx, w, r, nil, nil, err)")
	p(w, "			return")
	p(w, "		}")
	p(w, "")
	p(w, "		arg := &%s{}", ioname(m.GetInputType()))
	p(w, "")
	p(w, "		contentType := r.Header.Get(\"Content-Type\")")
	p(w, "		switch contentType {")
	p(w, "		case \"application/protobuf\", \"application/x-protobuf\":")
	p(w, "			if err := proto.Unmarshal(body, arg); err != nil {")
	p(w, "				cb(ctx, w, r, nil, nil, err)")
	p(w, "				return")
	p(w, "			}")
	p(w, "		case \"application/json\":")
	p(w, "			if err := jsonpb.Unmarshal(bytes.NewBuffer(body), arg); err != nil {")
	p(w, "				cb(ctx, w, r, nil, nil, err)")
	p(w, "				return")
	p(w, "			}")
	p(w, "		default:")
	p(w, "			w.WriteHeader(http.StatusUnsupportedMediaType)")
	p(w, "			_, err := fmt.Fprintf(w, \"Unsupported Content-Type: %%s\", contentType)")
	p(w, "			cb(ctx, w, r, nil, nil, err)")
	p(w, "			return")
	p(w, "		}")
	p(w, "")
	p(w, "		ret, err := h.srv.%s(ctx, arg)", m.GetName())
	p(w, "		if err != nil {")
	p(w, "			cb(ctx, w, r, arg, nil, err)")
	p(w, "			return")
	p(w, "		}")
	p(w, "")
	p(w, "		switch contentType {")
	p(w, "		case \"application/protobuf\", \"application/x-protobuf\":")
	p(w, "			buf, err := proto.Marshal(ret)")
	p(w, "			if err != nil {")
	p(w, "				cb(ctx, w, r, arg, ret, err)")
	p(w, "				return")
	p(w, "			}")
	p(w, "			if _, err := io.Copy(w, bytes.NewBuffer(buf)); err != nil {")
	p(w, "				cb(ctx, w, r, arg, ret, err)")
	p(w, "				return")
	p(w, "			}")
	p(w, "		case \"application/json\":")
	p(w, "			m := jsonpb.Marshaler{")
	p(w, "				EnumsAsInts:  true,")
	p(w, "				EmitDefaults: true,")
	p(w, "			}")
	p(w, "			if err := m.Marshal(w, ret); err != nil {")
	p(w, "				cb(ctx, w, r, arg, ret, err)")
	p(w, "				return")
	p(w, "			}")
	p(w, "		default:")
	p(w, "			w.WriteHeader(http.StatusUnsupportedMediaType)")
	p(w, "			_, err := fmt.Fprintf(w, \"Unsupported Content-Type: %%s\", contentType)")
	p(w, "			cb(ctx, w, r, arg, ret, err)")
	p(w, "			return")
	p(w, "		}")
	p(w, "		cb(ctx, w, r, arg, ret, nil)")
	p(w, "	})")
	p(w, "}")
	p(w, "")
}

func (g *Generator) writeMethodWithPath(w io.Writer, s *descriptor.ServiceDescriptorProto, m *descriptor.MethodDescriptorProto) {
	p(w, "func (h *%sHTTPConverter) %sWithName(cb func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error)) (string, string, http.HandlerFunc) {", s.GetName(), m.GetName())
	p(w, "	return \"%s\", \"%s\", h.%s(cb)", s.GetName(), m.GetName(), m.GetName())
	p(w, "}")
	p(w, "")
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

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			cb(ctx, w, r, nil, nil, err)
			return
		}

		arg := &{{ $method.Arg }}{}

		contentType := r.Header.Get("Content-Type")
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

		ret, err := h.srv.{{ $method.Name }}(ctx, arg)
		if err != nil {
			cb(ctx, w, r, arg, nil, err)
			return
		}

		switch contentType {
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
			_, err := fmt.Fprintf(w, "Unsupported Content-Type: %s", contentType)
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
