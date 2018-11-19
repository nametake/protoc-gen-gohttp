package example

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
)

func TestHelloWorldServer_SayHello(t *testing.T) {
	var tests = []struct {
		name    string
		reqFunc func() (*http.Request, error)
		cb      func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error)
		want    *HelloReply
	}{
		{
			name: "Content-Type JSON",
			reqFunc: func() (*http.Request, error) {
				p := &HelloRequest{
					Name: "John",
				}

				body := &bytes.Buffer{}
				if err := json.NewEncoder(body).Encode(p); err != nil {
					return nil, err
				}

				req := httptest.NewRequest(http.MethodPost, "/", body)
				req.Header.Set("Content-Type", "application/json")
				return req, nil
			},
			cb: nil,
			want: &HelloReply{
				Message: "Hello, John!",
			},
		},
		{
			name: "Content-Type Protobuf",
			reqFunc: func() (*http.Request, error) {
				p := &HelloRequest{
					Name: "Smith",
				}

				buf, err := proto.Marshal(p)
				if err != nil {
					return nil, err
				}
				body := bytes.NewBuffer(buf)

				req := httptest.NewRequest(http.MethodPost, "/", body)
				req.Header.Set("Content-Type", "application/protobuf")
				return req, nil
			},
			cb: nil,
			want: &HelloReply{
				Message: "Hello, Smith!",
			},
		},
	}

	handler := NewGreeterHandler(&HelloWorldServer{})

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req, err := tt.reqFunc()
			if err != nil {
				t.Fatal(err)
			}

			rec := httptest.NewRecorder()
			handler.SayHello(tt.cb).ServeHTTP(rec, req)

			resp := &HelloReply{}
			switch req.Header.Get("Content-Type") {
			case "application/protobuf":
				if err := proto.Unmarshal(rec.Body.Bytes(), resp); err != nil {
					t.Fatal(err)
				}
			case "application/json":
				if err := json.NewDecoder(rec.Body).Decode(resp); err != nil {
					t.Fatal(err)
				}
			default:
			}
			if diff := cmp.Diff(resp, tt.want); diff != "" {
				t.Errorf("%s", diff)
			}
		})
	}
}