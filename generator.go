package main

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

type Generator struct {
	files map[string]*bytes.Buffer
}

func New() *Generator {
	return &Generator{
		files: make(map[string]*bytes.Buffer),
	}
}

func (g Generator) P(name string, args ...string) {
	for _, arg := range args {
		if _, err := io.WriteString(g.files[name], arg); err != nil {
			panic(err)
		}
	}
	if _, err := io.WriteString(g.files[name], "\n"); err != nil {
		panic(err)
	}
}

func (g *Generator) Generate(req *plugin.CodeGeneratorRequest) (*plugin.CodeGeneratorResponse, error) {
	for _, f := range req.FileToGenerate {
		g.files[f] = &bytes.Buffer{}
	}

	for _, f := range req.ProtoFile {
		g.writePackage(f.GetName(), f)
		g.writeImports(f.GetName())
		for _, service := range f.GetService() {
			g.writeService(f.GetName(), service)
			for _, method := range service.GetMethod() {
				g.writeMethod(f.GetName(), service, method)
			}
		}
	}

	files := make([]*plugin.CodeGeneratorResponse_File, 0)
	for name, buf := range g.files {
		file := &plugin.CodeGeneratorResponse_File{
			Name:    proto.String(basename(name) + ".http.go"),
			Content: proto.String(buf.String()),
		}
		files = append(files, file)
	}
	return &plugin.CodeGeneratorResponse{
		File: files,
	}, nil
}

func (g *Generator) writePackage(name string, f *descriptor.FileDescriptorProto) {
	pkg := fmt.Sprintf("package %s", f.Options.GetGoPackage())
	g.P(name, pkg)
	g.P(name)
}

func (g *Generator) writeImports(name string) {
	g.P(name, "import (")
	g.P(name, "	\"bytes\"")
	g.P(name, "	\"context\"")
	g.P(name, "	\"encoding/json\"")
	g.P(name, "	\"fmt\"")
	g.P(name, "	\"io\"")
	g.P(name, "	\"io/ioutil\"")
	g.P(name, "	\"net/http\"")
	g.P(name, "")
	g.P(name, "	\"github.com/golang/protobuf/jsonpb\"")
	g.P(name, "	\"github.com/golang/protobuf/proto\"")
	g.P(name, ")")
	g.P(name)
}

func (g *Generator) writeService(name string, s *descriptor.ServiceDescriptorProto) {
	g.P(name, fmt.Sprintf("type %s struct{}", s.GetName()))
	g.P(name)

	g.P(name, fmt.Sprintf("func New%s() *%s {", s.GetName(), s.GetName()))
	g.P(name, fmt.Sprintf("	return &%s{}", s.GetName()))
	g.P(name, "}")
}

func (g *Generator) writeMethod(name string, s *descriptor.ServiceDescriptorProto, m *descriptor.MethodDescriptorProto) {
	g.P(name)
	g.P(name, fmt.Sprintf("func (g *%s) %s(srv %sServer, cb func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error)) http.HandlerFunc {", s.GetName(), m.GetName(), s.GetName()))
	g.P(name, "	if cb == nil {")
	g.P(name, "		cb = func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error) {")
	g.P(name, "			if err != nil {")
	g.P(name, "				w.WriteHeader(http.StatusInternalServerError)")
	g.P(name, "				fmt.Fprintf(w, \"%v: arg = %v: ret = %v\", err, arg, ret)")
	g.P(name, "			}")
	g.P(name, "		}")
	g.P(name, "	}")
	g.P(name, "	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {")
	g.P(name, "		ctx := r.Context()")
	g.P(name)
	g.P(name, "		body, err := ioutil.ReadAll(r.Body)")
	g.P(name, "		if err != nil {")
	g.P(name, "			cb(ctx, w, r, nil, nil, err)")
	g.P(name, "			return")
	g.P(name, "		}")
	g.P(name)
	g.P(name, fmt.Sprintf("		var arg *%s", ioname(m.GetInputType())))
	g.P(name)
	g.P(name, "		contentType := r.Header.Get(\"Content-Type\")")
	g.P(name, "		switch contentType {")
	g.P(name, "		case \"application/protobuf\", \"application/x-protobuf\":")
	g.P(name, "			if err := proto.Unmarshal(body, arg); err != nil {")
	g.P(name, "				cb(ctx, w, r, nil, nil, err)")
	g.P(name, "				return")
	g.P(name, "			}")
	g.P(name, "		case \"application/json\":")
	g.P(name, "			if err := jsonpb.Unmarshal(bytes.NewBuffer(body), arg); err != nil {")
	g.P(name, "				cb(ctx, w, r, nil, nil, err)")
	g.P(name, "				return")
	g.P(name, "			}")
	g.P(name, "		default:")
	g.P(name, "			w.WriteHeader(http.StatusUnsupportedMediaType)")
	g.P(name, "			if _, err := fmt.Fprintf(w, \"Unsupported Content-Type: %s\", contentType); err != nil {")
	g.P(name, "				cb(ctx, w, r, nil, nil, err)")
	g.P(name, "			}")
	g.P(name, "			return")
	g.P(name, "		}")
	g.P(name)
	g.P(name, fmt.Sprintf("		ret, err := srv.%s(ctx, arg)", m.GetName()))
	g.P(name, "		if err != nil {")
	g.P(name, "			cb(ctx, w, r, arg, ret, err)")
	g.P(name, "			return")
	g.P(name, "		}")
	g.P(name)
	g.P(name, "		switch contentType {")
	g.P(name, "		case \"application/protobuf\", \"application/x-protobuf\":")
	g.P(name, "			buf, err := proto.Marshal(ret)")
	g.P(name, "			if err != nil {")
	g.P(name, "				cb(ctx, w, r, arg, ret, err)")
	g.P(name, "				return")
	g.P(name, "			}")
	g.P(name, "			if _, err := io.Copy(w, bytes.NewBuffer(buf)); err != nil {")
	g.P(name, "				cb(ctx, w, r, arg, ret, err)")
	g.P(name, "				return")
	g.P(name, "			}")
	g.P(name, "		case \"application/json\":")
	g.P(name, "			if err := json.NewEncoder(w).Encode(ret); err != nil {")
	g.P(name, "				cb(ctx, w, r, arg, ret, err)")
	g.P(name, "				return")
	g.P(name, "			}")
	g.P(name, "		default:")
	g.P(name, "			w.WriteHeader(http.StatusUnsupportedMediaType)")
	g.P(name, "			if _, err := fmt.Fprintf(w, \"Unsupported Content-Type: %s\", contentType); err != nil {")
	g.P(name, "				cb(ctx, w, r, nil, nil, err)")
	g.P(name, "			}")
	g.P(name, "			return")
	g.P(name, "		}")
	g.P(name)
	g.P(name, "		cb(ctx, w, r, arg, ret, err)")
	g.P(name, "	})")
	g.P(name, "}")
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
