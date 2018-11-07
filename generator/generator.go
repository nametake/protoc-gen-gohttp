package generator

import (
	"io"

	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

type Generator struct {
	w io.Writer
}

func New() *Generator {
	return &Generator{}
}

func (g *Generator) P(args ...string) {
	for _, v := range args {
		io.WriteString(g.w, v)
	}
}

func (g *Generator) GenerateAllFiles() *plugin.CodeGeneratorResponse {
	return nil
}
