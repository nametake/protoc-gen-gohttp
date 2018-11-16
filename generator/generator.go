package generator

import (
	"io"

	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

type Generator struct {
	output io.Writer
}

func New() *Generator {
	return &Generator{}
}

func (g *Generator) P(args ...string) {
	for _, v := range args {
		io.WriteString(g.output, v)
	}
}

func (g *Generator) Generate(req *plugin.CodeGeneratorRequest) (*plugin.CodeGeneratorResponse, error) {
	return nil, nil
}
