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
	bufs := make(map[string]*bytes.Buffer)
	for _, f := range req.FileToGenerate {
		bufs[f] = &bytes.Buffer{}
	}

	for _, f := range req.ProtoFile {
		buf := bufs[f.GetName()]
		g.writePackage(buf, f)
		g.writeImports(buf)
		for _, service := range f.GetService() {
			g.writeService(buf, service)
			for _, method := range service.GetMethod() {
				g.writeMethod(buf, service, method)
			}
		}
	}

	files := make([]*plugin.CodeGeneratorResponse_File, 0)
	for name, buf := range bufs {
		content, err := format.Source(buf.Bytes())
		if err != nil {
			return nil, err
		}

		file := &plugin.CodeGeneratorResponse_File{
			Name:    proto.String(basename(name) + ".http.go"),
			Content: proto.String(string(content)),
		}
		files = append(files, file)
	}
	return &plugin.CodeGeneratorResponse{
		File: files,
	}, nil
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
	p(w, "type %s struct{}", s.GetName())
	p(w, "")

	p(w, "func New%s() *%s {", s.GetName(), s.GetName())
	p(w, "	return &%s{}", s.GetName())
	p(w, "}")
	p(w, "")
}

func (g *Generator) writeMethod(w io.Writer, s *descriptor.ServiceDescriptorProto, m *descriptor.MethodDescriptorProto) {
	p(w, "func (g *%s) %s(srv %sServer, cb func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error)) http.HandlerFunc {", s.GetName(), m.GetName(), s.GetName())
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
	p(w, "		var arg *%s", ioname(m.GetInputType()))
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
	p(w, "			if _, err := fmt.Fprintf(w, \"Unsupported Content-Type: %%s\", contentType); err != nil {")
	p(w, "				cb(ctx, w, r, nil, nil, err)")
	p(w, "			}")
	p(w, "			return")
	p(w, "		}")
	p(w, "")
	p(w, "		ret, err := srv.%s(ctx, arg)", m.GetName())
	p(w, "		if err != nil {")
	p(w, "			cb(ctx, w, r, arg, ret, err)")
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
	p(w, "			if _, err := fmt.Fprintf(w, \"Unsupported Content-Type: %%s\", contentType); err != nil {")
	p(w, "				cb(ctx, w, r, nil, nil, err)")
	p(w, "			}")
	p(w, "			return")
	p(w, "		}")
	p(w, "")
	p(w, "		cb(ctx, w, r, arg, ret, err)")
	p(w, "	})")
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
