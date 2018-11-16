package generator

import (
	"bytes"
	"fmt"
	"io"
	"path"

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

func basename(name string) string {
	if ext := path.Ext(name); ext == ".proto" || ext == ".protodevel" {
		name = name[:len(name)-len(ext)]
	}
	return name
}
