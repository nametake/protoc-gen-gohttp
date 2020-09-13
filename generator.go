package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var toCamelCaseRe = regexp.MustCompile(`(^[A-Za-z])|(_|\.)([A-Za-z])`)

func toCamelCase(str string) string {
	return toCamelCaseRe.ReplaceAllStringFunc(str, func(s string) string {
		return strings.ToUpper(strings.Replace(s, "_", "", -1))
	})
}

type pathParam struct {
	Index  int
	Name   string
	GoName string
}

func (t *pathParam) GetSplitedGoNames() []string {
	names := make([]string, 0)
	ps := strings.Split(t.GoName, ".")
	for i := range ps {
		if i == 0 {
			continue
		}
		names = append(names, strings.Join(ps[0:i], "."))
	}
	return names
}

func parsePathParam(pattern string) ([]*pathParam, error) {
	if !strings.HasPrefix(pattern, "/") {
		return nil, fmt.Errorf("no leading /")
	}
	tokens, _ := tokenize(pattern[1:])

	p := parser{tokens: tokens}
	segs, err := p.topLevelSegments()
	if err != nil {
		return nil, err
	}

	params := make([]*pathParam, 0)
	for i, seg := range segs {
		if v, ok := seg.(variable); ok {
			params = append(params, &pathParam{
				Index:  i + 1,
				Name:   v.path,
				GoName: toCamelCase(v.path),
			})
		}
	}

	sort.Slice(params, func(i, j int) bool {
		a := params[i]
		b := params[j]
		if len(strings.Split(a.Name, ".")) < len(strings.Split(b.Name, ".")) {
			return true
		}
		return params[i].Name < params[j].Name
	})

	return params, nil
}

type queryParam struct {
	*protogen.Field

	GoName string
	Name   string
}

func createQueryParams(method *protogen.Method) []*queryParam {
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
