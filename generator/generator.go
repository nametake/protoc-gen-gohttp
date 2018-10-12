package generator

import "io"

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
