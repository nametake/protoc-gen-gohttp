package generator

import (
	"io"
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

func TestGenerator_GenerateAllFiles(t *testing.T) {
	type fields struct {
		w io.Writer
	}
	tests := []struct {
		name   string
		fields fields
		want   *plugin.CodeGeneratorResponse
	}{
		{
			name: "helloworld",
			fields: fields{
				w: nil,
			},
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
			g := &Generator{
				w: tt.fields.w,
			}
			if got := g.GenerateAllFiles(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Generator.GenerateAllFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}
