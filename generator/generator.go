package generator

import (
	"bytes"
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

func (g *Generator) Generate(req *plugin.CodeGeneratorRequest) (*plugin.CodeGeneratorResponse, error) {
	for _, f := range req.FileToGenerate {
		g.files[f] = &bytes.Buffer{}
	}

	files := make([]*plugin.CodeGeneratorResponse_File, 0)
	for name := range g.files {
		file := &plugin.CodeGeneratorResponse_File{
			Name:    proto.String(basename(name) + ".http.go"),
			Content: proto.String("package gohttp\n"),
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
