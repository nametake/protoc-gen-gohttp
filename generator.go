package main

import (
	"bytes"
	"go/format"
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/pseudomuto/protokit"
	"google.golang.org/genproto/googleapis/api/annotations"
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
	Name     string
	Arg      string
	Comment  string
	HTTPRule *targetHTTPRule
}

type targetHTTPRule struct {
	Method    string
	Pattern   string
	Valiables []*targetValiable
}

func (t *targetHTTPRule) GetMethod() string {
	switch t.Method {
	case http.MethodGet:
		return "http.MethodGet"
	case http.MethodPut:
		return "http.MethodPut"
	case http.MethodPost:
		return "http.MethodPost"
	case http.MethodDelete:
		return "http.MethodDelete"
	case http.MethodPatch:
		return "http.MethodPatch"
	}
	return ""
}

type targetValiable struct {
	Index int
	Path  string
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
				Name:     method.GetName(),
				Arg:      ioname(method.GetInputType()),
				Comment:  method.GetComments().GetLeading(),
				HTTPRule: parseHTTPRule(method),
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

func parseHTTPRule(md *protokit.MethodDescriptor) *targetHTTPRule {
	if httpRule, ok := md.OptionExtensions["google.api.http"].(*annotations.HttpRule); ok {
		target := &targetHTTPRule{}
		switch httpRule.GetPattern().(type) {
		case *annotations.HttpRule_Get:
			target.Method = http.MethodGet
			target.Pattern = httpRule.GetGet()
		case *annotations.HttpRule_Put:
			target.Method = http.MethodPut
			target.Pattern = httpRule.GetPut()
		case *annotations.HttpRule_Post:
			target.Method = http.MethodPost
			target.Pattern = httpRule.GetPost()
		case *annotations.HttpRule_Delete:
			target.Method = http.MethodDelete
			target.Pattern = httpRule.GetDelete()
		case *annotations.HttpRule_Patch:
			target.Method = http.MethodPatch
			target.Pattern = httpRule.GetPatch()
		}
		return target
	}
	return nil
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
