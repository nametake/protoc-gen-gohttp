package main

import (
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type targetVariable struct {
	Index int
	Path  string
}

func (t *targetVariable) GetPath() string {
	return toCamelCase(t.Path)
}

func (t *targetVariable) GetPaths() []string {
	paths := make([]string, 0)
	ps := strings.Split(t.Path, ".")
	for i := range ps {
		if i == 0 {
			continue
		}
		paths = append(paths, toCamelCase(strings.Join(ps[0:i], ".")))
	}
	return paths
}

type queryParam struct {
	*protogen.Field

	GoName string
	Name   string
}

func parseQueryParam(method *protogen.Method) []*queryParam {
	queryParams := make([]*queryParam, 0)

	var f func(parent *queryParam, fields []*protogen.Field)

	f = func(parent *queryParam, fields []*protogen.Field) {
		for _, field := range fields {
			if field.Desc.Kind() == protoreflect.MessageKind {
				q := &queryParam{
					Field:  field,
					GoName: fmt.Sprintf("%s.", field.GoName),
					Name:   fmt.Sprintf("%s.", field.Desc.Name()),
				}
				f(q, field.Message.Fields)
				continue
			}
			queryParams = append(queryParams, &queryParam{
				Field:  field,
				GoName: fmt.Sprintf("%s%s", parent.GoName, field.GoName),
				Name:   fmt.Sprintf("%s%s", parent.Name, field.Desc.Name()),
			})
		}
	}

	f(&queryParam{GoName: "", Name: ""}, method.Input.Fields)

	return queryParams
}

var toCamelCaseRe = regexp.MustCompile(`(^[A-Za-z])|(_|\.)([A-Za-z])`)

func toCamelCase(str string) string {
	return toCamelCaseRe.ReplaceAllStringFunc(str, func(s string) string {
		return strings.ToUpper(strings.Replace(s, "_", "", -1))
	})
}
