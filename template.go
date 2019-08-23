package main

const codeTemplate = `// Code generated by protoc-gen-gohttp. DO NOT EDIT.
// source: {{ .Name }}

package {{ .Pkg }}

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)
{{ range $i, $service := .Services }}
// {{ $service.Name }}HTTPConverter has a function to convert {{ $service.Name }}Server interface to http.HandlerFunc.
type {{ $service.Name }}HTTPConverter struct {
	srv {{ $service.Name }}Server
}

// New{{ $service.Name }}HTTPConverter returns {{ $service.Name }}HTTPConverter.
func New{{ $service.Name }}HTTPConverter(srv {{ $service.Name }}Server) *{{ $service.Name }}HTTPConverter {
	return &{{ $service.Name }}HTTPConverter{
		srv: srv,
	}
}
{{ range $j, $method := $service.Methods }}
// {{ $method.Name }} returns {{ $service.Name }}Server interface's {{ $method.Name }} converted to http.HandlerFunc.
{{ if ne $method.Comment "" -}}
//
// {{ $method.Comment }}
{{ end -}}
func (h *{{ $service.Name }}HTTPConverter) {{ $method.Name }}(cb func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error)) http.HandlerFunc {
	if cb == nil {
		cb = func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error) {
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				p := status.New(codes.Unknown, err.Error()).Proto()
				switch r.Header.Get("Content-Type") {
				case "application/protobuf", "application/x-protobuf":
					buf, err := proto.Marshal(p)
					if err != nil {
						return
					}
					if _, err := io.Copy(w, bytes.NewBuffer(buf)); err != nil {
						return
					}
				case "application/json":
					if err := json.NewEncoder(w).Encode(p); err != nil {
						return
					}
				default:
				}
			}
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		arg := &{{ $method.Arg }}{}
		contentType := r.Header.Get("Content-Type")
		if r.Method != http.MethodGet {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				cb(ctx, w, r, nil, nil, err)
				return
			}

			switch contentType {
			case "application/protobuf", "application/x-protobuf":
				if err := proto.Unmarshal(body, arg); err != nil {
					cb(ctx, w, r, nil, nil, err)
					return
				}
			case "application/json":
				if err := jsonpb.Unmarshal(bytes.NewBuffer(body), arg); err != nil {
					cb(ctx, w, r, nil, nil, err)
					return
				}
			default:
				w.WriteHeader(http.StatusUnsupportedMediaType)
				_, err := fmt.Fprintf(w, "Unsupported Content-Type: %s", contentType)
				cb(ctx, w, r, nil, nil, err)
				return
			}
		}

		ret, err := h.srv.{{ $method.Name }}(ctx, arg)
		if err != nil {
			cb(ctx, w, r, arg, nil, err)
			return
		}

		accepts := strings.Split(r.Header.Get("Accept"), ",")
		accept := accepts[0]
		if accept == "*/*" || accept == ""{
			if contentType != "" {
				accept = contentType
			} else {
				accept = "application/json"
			}
		}

		switch accept {
		case "application/protobuf", "application/x-protobuf":
			buf, err := proto.Marshal(ret)
			if err != nil {
				cb(ctx, w, r, arg, ret, err)
				return
			}
			if _, err := io.Copy(w, bytes.NewBuffer(buf)); err != nil {
				cb(ctx, w, r, arg, ret, err)
				return
			}
		case "application/json":
			m := jsonpb.Marshaler{
				EnumsAsInts:  true,
				EmitDefaults: true,
			}
			if err := m.Marshal(w, ret); err != nil {
				cb(ctx, w, r, arg, ret, err)
				return
			}
		default:
			w.WriteHeader(http.StatusUnsupportedMediaType)
			_, err := fmt.Fprintf(w, "Unsupported Accept: %s", accept)
			cb(ctx, w, r, arg, ret, err)
			return
		}
		cb(ctx, w, r, arg, ret, nil)
	})
}

// {{ $method.Name }}WithName returns Service name, Method name and {{ $service.Name }}Server interface's {{ $method.Name }} converted to http.HandlerFunc.
{{ if ne $method.Comment "" -}}
//
// {{ $method.Comment }}
{{ end -}}
func (h *{{ $service.Name }}HTTPConverter) {{ $method.Name }}WithName(cb func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error)) (string, string, http.HandlerFunc) {
	return "{{ $service.Name }}", "{{ $method.Name }}", h.{{ $method.Name }}(cb)
}

{{ if $method.HTTPRule -}}
func (h *{{ $service.Name }}HTTPConverter) {{ $method.Name }}HTTPRule(cb func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error)) (string, string, http.HandlerFunc) {
	if cb == nil {
		cb = func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error) {
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				p := status.New(codes.Unknown, err.Error()).Proto()
				switch r.Header.Get("Content-Type") {
				case "application/protobuf", "application/x-protobuf":
					buf, err := proto.Marshal(p)
					if err != nil {
						return
					}
					if _, err := io.Copy(w, bytes.NewBuffer(buf)); err != nil {
						return
					}
				case "application/json":
					if err := json.NewEncoder(w).Encode(p); err != nil {
						return
					}
				default:
				}
			}
		}
	}
	return {{ $method.HTTPRule.GetMethod }}, "{{ $method.HTTPRule.Pattern }}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		arg := &{{ $method.Arg }}{}
		contentType := r.Header.Get("Content-Type")
		{{ if $method.GetQueryParams -}}
		if r.Method == http.MethodGet {
			var err error
		{{ range $k, $queryParam := $method.GetQueryParams -}}
		{{ template "queryString" $queryParam -}}
		{{ end -}}
		}
		{{ else -}}
		if r.Method != http.MethodGet {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				cb(ctx, w, r, nil, nil, err)
				return
			}

			switch contentType {
			case "application/protobuf", "application/x-protobuf":
				if err := proto.Unmarshal(body, arg); err != nil {
					cb(ctx, w, r, nil, nil, err)
					return
				}
			case "application/json":
				if err := jsonpb.Unmarshal(bytes.NewBuffer(body), arg); err != nil {
					cb(ctx, w, r, nil, nil, err)
					return
				}
			default:
				w.WriteHeader(http.StatusUnsupportedMediaType)
				_, err := fmt.Fprintf(w, "Unsupported Content-Type: %s", contentType)
				cb(ctx, w, r, nil, nil, err)
				return
			}
		}
		{{ end }}

		{{ if $method.HTTPRule.Variables -}}
		p := strings.Split(r.URL.Path, "/")
		{{ range $j, $variable := $method.HTTPRule.Variables -}}
		arg.{{ $variable.GetPath }} = p[{{ $variable.Index }}]
		{{ end -}}
		{{ end }}

		ret, err := h.srv.{{ $method.Name }}(ctx, arg)
		if err != nil {
			cb(ctx, w, r, arg, nil, err)
			return
		}

		accepts := strings.Split(r.Header.Get("Accept"), ",")
		accept := accepts[0]
		if accept == "*/*" || accept == ""{
			if contentType != "" {
				accept = contentType
			} else {
				accept = "application/json"
			}
		}

		switch accept {
		case "application/protobuf", "application/x-protobuf":
			buf, err := proto.Marshal(ret)
			if err != nil {
				cb(ctx, w, r, arg, ret, err)
				return
			}
			if _, err := io.Copy(w, bytes.NewBuffer(buf)); err != nil {
				cb(ctx, w, r, arg, ret, err)
				return
			}
		case "application/json":
			m := jsonpb.Marshaler{
				EnumsAsInts:  true,
				EmitDefaults: true,
			}
			if err := m.Marshal(w, ret); err != nil {
				cb(ctx, w, r, arg, ret, err)
				return
			}
		default:
			w.WriteHeader(http.StatusUnsupportedMediaType)
			_, err := fmt.Fprintf(w, "Unsupported Accept: %s", accept)
			cb(ctx, w, r, arg, ret, err)
			return
		}
		cb(ctx, w, r, arg, ret, nil)
	})
}
{{ end -}}
{{ end -}}
{{ end -}}
` + queryParamsTemplate

const queryParamsTemplate = `{{ define "queryString" -}}
{{ if eq .QueryType "STRING" -}}
arg.{{ .GetPath }} = r.URL.Query().Get("{{ .Key }}")
{{ else if eq .QueryType "INT64" -}}
arg.{{ .GetPath }}, err = strconv.ParseInt(r.URL.Query().Get("{{ .Key }}"), 10, 64)
if err != nil {
	cb(ctx, w, r, nil, nil, err)
	return
}
{{ end -}}
{{ end -}}`
