package main

import (
	"bytes"
	"fmt"
	"go/format"
	"html/template"
	"log"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/pseudomuto/protokit"
	"google.golang.org/genproto/googleapis/api/annotations"
)

type targetFile struct {
	Name     string
	Pkg      string
	Services []*targetService
}

func (t *targetFile) IsImportStrConv() bool {
	for _, service := range t.Services {
		for _, method := range service.Methods {
			for _, queryParam := range method.QueryParams {
				switch queryParam.QueryType {
				case queryInt64, queryString:
					return true
				}
			}
		}
	}
	return false
}

type targetService struct {
	Name    string
	Methods []*targetMethod
}

type targetMethod struct {
	Name        string
	Arg         string
	Comment     string
	HTTPRule    *targetHTTPRule
	QueryParams []*targetQueryParam
}

func (t *targetMethod) GetQueryParams() []*targetQueryParam {
	params := make([]*targetQueryParam, 0)
	for _, param := range t.QueryParams {
		for _, v := range t.HTTPRule.Variables {
			if param.Path == v.Path {
				continue
			}
			params = append(params, param)
		}
	}
	return params
}

type targetHTTPRule struct {
	Method    string
	Pattern   string
	Variables []*targetVariable
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

type targetVariable struct {
	Index int
	Path  string
}

func (t *targetVariable) GetPath() string {
	return toCamelCase(t.Path)
}

const (
	queryString = "STRING"
	queryInt64  = "INT64"
)

type targetQueryParam struct {
	QueryType string
	Path      string
}

func (t *targetQueryParam) GetPath() string {
	return toCamelCase(t.Path)
}

func (t *targetQueryParam) Key() string {
	return t.Path
}

// Generate receives a CodeGeneratorRequest and returns a CodeGeneratorResponse.
func Generate(req *plugin.CodeGeneratorRequest) (*plugin.CodeGeneratorResponse, error) {
	descriptors := protokit.ParseCodeGenRequest(req)

	// Filter files to target files.
	targetFiles := make([]*targetFile, 0)
	for _, f := range descriptors {
		target, err := genTarget(f)
		if err != nil {
			return nil, err
		}
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

func genTarget(file *protokit.FileDescriptor) (*targetFile, error) {
	if len(file.GetServices()) == 0 {
		return nil, nil
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
			httpRule, err := parseHTTPRule(method)
			if err != nil {
				return nil, err
			}

			s.Methods = append(s.Methods, &targetMethod{
				Name:        method.GetName(),
				Arg:         ioname(method.GetInputType()),
				Comment:     method.GetComments().GetLeading(),
				HTTPRule:    httpRule,
				QueryParams: parseQueryParam(method, file.GetMessages()),
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
		return nil, nil
	}
	return f, nil
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

func parseHTTPRule(md *protokit.MethodDescriptor) (*targetHTTPRule, error) {
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

		valiables, err := parsePattern(target.Pattern)
		if err != nil {
			return nil, err
		}
		target.Variables = valiables

		return target, nil
	}
	return nil, nil
}

func parseQueryParam(md *protokit.MethodDescriptor, msgs []*protokit.Descriptor) []*targetQueryParam {
	httpRule, ok := md.OptionExtensions["google.api.http"].(*annotations.HttpRule)
	if !ok {
		return nil
	}
	if _, ok := httpRule.GetPattern().(*annotations.HttpRule_Get); !ok {
		return nil
	}
	queryParams := make([]*targetQueryParam, 0)

	// Define func that to parse field name and type recursively.
	var f func(parent string, fields []*protokit.FieldDescriptor)
	f = func(parent string, fields []*protokit.FieldDescriptor) {
		for _, field := range fields {
			label := field.GetLabel()
			typ := field.GetType()
			switch {
			case label == descriptor.FieldDescriptorProto_LABEL_OPTIONAL && typ == descriptor.FieldDescriptorProto_TYPE_INT64:
				queryParams = append(queryParams, &targetQueryParam{
					QueryType: queryInt64,
					Path:      fmt.Sprintf("%s%s", parent, field.GetName()),
				})
			case label == descriptor.FieldDescriptorProto_LABEL_OPTIONAL && typ == descriptor.FieldDescriptorProto_TYPE_STRING:
				queryParams = append(queryParams, &targetQueryParam{
					QueryType: queryString,
					Path:      fmt.Sprintf("%s%s", parent, field.GetName()),
				})
			case label == descriptor.FieldDescriptorProto_LABEL_OPTIONAL && typ == descriptor.FieldDescriptorProto_TYPE_MESSAGE:
				for _, msg := range msgs {
					if strings.HasSuffix(field.GetTypeName(), msg.GetFullName()) {
						f(fmt.Sprintf("%s.", field.GetName()), msg.GetMessageFields())
						break
					} else if strings.Contains(field.GetTypeName(), msg.GetFullName()) {
						for _, m := range msg.GetMessages() {
							f(fmt.Sprintf("%s.", field.GetName()), m.GetMessageFields())
						}
					}
				}
			}
		}
	}

	var input *protokit.Descriptor
	for _, msg := range msgs {
		if strings.HasSuffix(md.GetInputType(), msg.GetFullName()) {
			input = msg
			break
		}
	}

	if input == nil {
		return nil
	}

	f("", input.GetMessageFields())

	return queryParams
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

var toCamelCaseRe = regexp.MustCompile(`(^[A-Za-z])|(_|\.)([A-Za-z])`)

func toCamelCase(str string) string {
	return toCamelCaseRe.ReplaceAllStringFunc(str, func(s string) string {
		return strings.ToUpper(strings.Replace(s, "_", "", -1))
	})
}
