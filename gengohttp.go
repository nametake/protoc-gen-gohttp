package main

import (
	"google.golang.org/protobuf/compiler/protogen"
)

func GenerateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	filename := file.GeneratedFilenamePrefix + ".http.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)

	g.P("// comment")
	g.P("package main")

	return g
}
