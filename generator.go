package main

import (
	"bytes"
	"go/format"
	"html/template"
	"log"
	"path"
	"strings"

	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/pseudomuto/protokit"
)

type targetFile struct {
	Name     string
	Pkg      string
	Services []*targetService
}

type targetService struct {
	Name    string
	Methods []*targetMethod
}

type targetMethod struct {
	Name    string
	Arg     string
	Comment string
}

// Generate receives a CodeGeneratorRequest and returns a CodeGeneratorResponse.
func Generate(req *plugin.CodeGeneratorRequest) (*plugin.CodeGeneratorResponse, error) {
	descriptors := protokit.ParseCodeGenRequest(req)

	// Filter files to target files.
	targetFiles := make([]*targetFile, 0)
	for _, f := range descriptors {
		target := genTarget(f)
		if target != nil {
			targetFiles = append(targetFiles, target)
		}
	}

	// Generate response files from proto files.
	respFiles := make([]*plugin.CodeGeneratorResponse_File, 0)
	for _, f := range targetFiles {
		respFiles = append(respFiles, genRespFile(f))
	}

	// Format response files content.
	for _, f := range respFiles {
		content, err := format.Source([]byte(f.GetContent()))
		if err != nil {
			return nil, err
		}
		f.Content = proto.String(string(content))
	}

	return &plugin.CodeGeneratorResponse{
		File: respFiles,
	}, nil
}

func genTarget(file *protokit.FileDescriptor) *targetFile {
	if len(file.GetServices()) == 0 {
		return nil
	}
	f := &targetFile{
		Name:     file.GetName(),
		Pkg:      file.GetOptions().GetGoPackage(),
		Services: make([]*targetService, 0),
	}

	for _, service := range file.GetServices() {
		s := &targetService{
			Name:    service.GetName(),
			Methods: make([]*targetMethod, 0),
		}
		for _, method := range service.GetMethods() {
			// Not generate streaming method
			if method.GetServerStreaming() || method.GetClientStreaming() {
				continue
			}
			s.Methods = append(s.Methods, &targetMethod{
				Name:    method.GetName(),
				Arg:     ioname(method.GetInputType()),
				Comment: method.GetComments().GetLeading(),
			})
		}
		// Add if Service has a method
		if len(s.Methods) == 0 {
			continue
		}
		f.Services = append(f.Services, s)
	}

	// Generate if File has a service
	if len(f.Services) == 0 {
		return nil
	}
	return f
}

func genRespFile(target *targetFile) *plugin.CodeGeneratorResponse_File {
	buf := &bytes.Buffer{}

	t := template.Must(template.New("gohttp").Parse(codeTemplate))
	if err := t.Execute(buf, target); err != nil {
		log.Println("executing template:", err)
		panic(err)
	}

	return &plugin.CodeGeneratorResponse_File{
		Name:    proto.String(basename(target.Name) + ".http.go"),
		Content: proto.String(buf.String()),
	}
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
