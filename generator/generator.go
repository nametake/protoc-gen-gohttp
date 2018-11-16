package generator

import (
	"bytes"
	"fmt"
	"io"
	"path"

	"github.com/gogo/protobuf/proto"
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
		pkg := fmt.Sprintf("package %s", f.Options.GetGoPackage())

		g.P(f.GetName(), pkg)
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

func basename(name string) string {
	if ext := path.Ext(name); ext == ".proto" || ext == ".protodevel" {
		name = name[:len(name)-len(ext)]
	}
	return name
}
