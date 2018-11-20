package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
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
	file     *descriptor.FileDescriptorProto
	services []*targetService
}

type targetService struct {
	service *descriptor.ServiceDescriptorProto
	methods []*descriptor.MethodDescriptorProto
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
		file: file,
	}

	tService := &targetService{}
	for _, s := range file.GetService() {
		tService.service = s
		for _, m := range s.GetMethod() {
			if !m.GetServerStreaming() && !m.GetClientStreaming() {
				tService.methods = append(tService.methods, m)
			}
		}
		if len(tService.methods) > 0 {
			tFile.services = append(tFile.services, tService)
		}
	}

	if len(tFile.services) > 0 {
		return tFile
	}

	return nil
}

func (g *Generator) genRespFile(target *targetFile) *plugin.CodeGeneratorResponse_File {
	buf := &bytes.Buffer{}
	g.writePackage(buf, target.file)
	g.writeImports(buf)
	for _, service := range target.services {
		g.writeService(buf, service.service)
		for _, method := range service.methods {
			g.writeMethod(buf, service.service, method)
			g.writeMethodWithPath(buf, service.service, method)
		}
	}
	return &plugin.CodeGeneratorResponse_File{
		Name:    proto.String(basename(target.file.GetName()) + ".http.go"),
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
	p(w, "				fmt.Fprintf(w, \"%%v: arg = %%v: ret = %%v\", err, arg, ret)")
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
	p(w, "			if err := json.NewEncoder(w).Encode(ret); err != nil {")
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
	p(w, "func (h *%sHTTPConverter) %sWithPath(cb func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error)) (string, http.HandlerFunc) {", s.GetName(), m.GetName())
	p(w, "	return \"/%s/%s\", h.%s(cb)", strings.ToLower(s.GetName()), strings.ToLower(m.GetName()), m.GetName())
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
