package generator

import (
	"fmt"
	"os"
	"testing"

	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/google/go-cmp/cmp"
)

func TestGenerator_GenerateAllFiles(t *testing.T) {
	type fields struct {
	}
	tests := []struct {
		name   string
		fields fields
		want   *plugin.CodeGeneratorResponse
	}{
		{
			name:   "helloworld",
			fields: fields{},
			want: &plugin.CodeGeneratorResponse{
				File: []*plugin.CodeGeneratorResponse_File{
					{
						Name:    proto.String("foo"),
						Content: proto.String("bar"),
					},
				},
			},
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
			if diff := cmp.Diff(g.GenerateAllFiles(), tt.want); diff != "" {
				t.Errorf("%s", diff)
			}
		})
	}
}
