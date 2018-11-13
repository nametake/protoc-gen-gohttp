package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/google/go-cmp/cmp"
)

func TestGenerator_GenerateAllFiles(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "helloworld",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := os.Open(fmt.Sprintf("testdata/%s.proto", tt.name))
			if err != nil {
				t.Fatalf("failed to open proto file: %v", err)
			}
			g := &Generator{
				w: p,
			}

			content, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s.http.go", tt.name))
			if err != nil {
				t.Fatalf("failed to read want file: %v", err)
			}

			want := &plugin.CodeGeneratorResponse{
				File: []*plugin.CodeGeneratorResponse_File{
					{
						Name:    proto.String(fmt.Sprintf("%s.http.go", tt.name)),
						Content: proto.String(string(content)),
					},
				},
			}

			if diff := cmp.Diff(g.GenerateAllFiles(), want); diff != "" {
				t.Errorf("%s", diff)
			}
		})
	}
}
