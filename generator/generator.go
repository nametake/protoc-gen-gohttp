package generator

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
	g.P(name, "\t\"bytes\"")
	g.P(name, "\t\"context\"")
	g.P(name, "\t\"encoding/json\"")
	g.P(name, "\t\"fmt\"")
	g.P(name, "\t\"io\"")
	g.P(name, "\t\"io/ioutil\"")
	g.P(name, "\t\"net/http\"")
	g.P(name, "")
	g.P(name, "\t\"github.com/golang/protobuf/jsonpb\"")
	g.P(name, "\t\"github.com/golang/protobuf/proto\"")
	g.P(name, ")")
	g.P(name)
}

func (g *Generator) writeService(name string, s *descriptor.ServiceDescriptorProto) {
	g.P(name, fmt.Sprintf("type %s struct{}", s.GetName()))
	g.P(name)

	g.P(name, fmt.Sprintf("func New%s() *%s {", s.GetName(), s.GetName()))
	g.P(name, fmt.Sprintf("\treturn &%s{}", s.GetName()))
	g.P(name, "}")
	g.P(name)
}

func (g *Generator) writeMethod(name string, s *descriptor.ServiceDescriptorProto, m *descriptor.MethodDescriptorProto) {
	g.P(name, fmt.Sprintf("func (g *%s) %s(srv %sServer, cb func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error)) http.HandlerFunc {", s.GetName(), m.GetName(), s.GetName()))
	g.P(name, "\tif cb == nil {")
	g.P(name, "\t\tcb = func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error) {")
	g.P(name, "\t\t\tif err != nil {")
	g.P(name, "\t\t\t\tw.WriteHeader(http.StatusInternalServerError)")
	g.P(name, "\t\t\t\tfmt.Fprintf(w, \"%v: arg = %v: ret = %v\", err, arg, ret)")
	g.P(name, "\t\t\t}")
	g.P(name, "\t\t}")
	g.P(name, "\t}")
	g.P(name, "\treturn http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {")
	g.P(name, "\t\tctx := r.Context()")
	g.P(name)
	g.P(name, "\t\tbody, err := ioutil.ReadAll(r.Body)")
	g.P(name, "\t\tif err != nil {")
	g.P(name, "\t\t\tcb(ctx, w, r, nil, nil, err)")
	g.P(name, "\t\t\treturn")
	g.P(name, "\t\t}")
	g.P(name)
	g.P(name, fmt.Sprintf("\t\tvar arg *%s", ioname(m.GetInputType())))
	g.P(name)
	g.P(name, "\t\tcontentType := r.Header.Get(\"Content-Type\")")
	g.P(name, "\t\tswitch contentType {")
	g.P(name, "\t\tcase \"application/protobuf\", \"application/x-protobuf\":")
	g.P(name, "\t\t\tif err := proto.Unmarshal(body, arg); err != nil {")
	g.P(name, "\t\t\t\tcb(ctx, w, r, nil, nil, err)")
	g.P(name, "\t\t\t\treturn")
	g.P(name, "\t\t\t}")
	g.P(name, "\t\tcase \"application/json\":")
	g.P(name, "\t\t\tif err := jsonpb.Unmarshal(bytes.NewBuffer(body), arg); err != nil {")
	g.P(name, "\t\t\t\tcb(ctx, w, r, nil, nil, err)")
	g.P(name, "\t\t\t\treturn")
	g.P(name, "\t\t\t}")
	g.P(name, "\t\tdefault:")
	g.P(name, "\t\t\tw.WriteHeader(http.StatusUnsupportedMediaType)")
	g.P(name, "\t\t\tif _, err := fmt.Fprintf(w, \"Unsupported Content-Type: %s\", contentType); err != nil {")
	g.P(name, "\t\t\t\tcb(ctx, w, r, nil, nil, err)")
	g.P(name, "\t\t\t}")
	g.P(name, "\t\t\treturn")
	g.P(name, "\t\t}")
	g.P(name)
	g.P(name, fmt.Sprintf("\t\tret, err := srv.%s(ctx, arg)", m.GetName()))
	g.P(name, "\t\tif err != nil {")
	g.P(name, "\t\t\tcb(ctx, w, r, arg, ret, err)")
	g.P(name, "\t\t\treturn")
	g.P(name, "\t\t}")
	g.P(name)
	g.P(name, "\t\tswitch contentType {")
	g.P(name, "\t\tcase \"application/protobuf\", \"application/x-protobuf\":")
	g.P(name, "\t\t\tbuf, err := proto.Marshal(ret)")
	g.P(name, "\t\t\tif err != nil {")
	g.P(name, "\t\t\t\tcb(ctx, w, r, arg, ret, err)")
	g.P(name, "\t\t\t\treturn")
	g.P(name, "\t\t\t}")
	g.P(name, "\t\t\tif _, err := io.Copy(w, bytes.NewBuffer(buf)); err != nil {")
	g.P(name, "\t\t\t\tcb(ctx, w, r, arg, ret, err)")
	g.P(name, "\t\t\t\treturn")
	g.P(name, "\t\t\t}")
	g.P(name, "\t\tcase \"application/json\":")
	g.P(name, "\t\t\tif err := json.NewEncoder(w).Encode(ret); err != nil {")
	g.P(name, "\t\t\t\tcb(ctx, w, r, arg, ret, err)")
	g.P(name, "\t\t\t\treturn")
	g.P(name, "\t\t\t}")
	g.P(name, "\t\tdefault:")
	g.P(name, "\t\t\tw.WriteHeader(http.StatusUnsupportedMediaType)")
	g.P(name, "\t\t\tif _, err := fmt.Fprintf(w, \"Unsupported Content-Type: %s\", contentType); err != nil {")
	g.P(name, "\t\t\t\tcb(ctx, w, r, nil, nil, err)")
	g.P(name, "\t\t\t}")
	g.P(name, "\t\t\treturn")
	g.P(name, "\t\t}")
	g.P(name)
	g.P(name, "\t\tcb(ctx, w, r, arg, ret, err)")
	g.P(name, "\t})")
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
