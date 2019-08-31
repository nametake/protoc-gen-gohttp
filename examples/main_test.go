package main

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

func TestEchoGreeterServer_SayHello(t *testing.T) {
	type want struct {
		StatusCode int
		Resp       *HelloReply
	}
	var tests = []struct {
		name    string
		reqFunc func() (*http.Request, error)
		cb      func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error)
		wantErr bool
		want    *want
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
			cb:      nil,
			wantErr: false,
			want: &want{
				StatusCode: 200,
				Resp: &HelloReply{
					Message: "Hello, John!",
				},
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
			cb:      nil,
			wantErr: false,
			want: &want{
				StatusCode: 200,
				Resp: &HelloReply{
					Message: "Hello, Smith!",
				},
			},
		},
		{
			name: "Nil body JSON",
			reqFunc: func() (*http.Request, error) {
				req := httptest.NewRequest(http.MethodPost, "/", nil)
				req.Header.Set("Content-Type", "application/json")
				return req, nil
			},
			cb: func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error) {
				if arg != nil {
					t.Errorf("arg is not nil: %#v", arg)
				}
				if ret != nil {
					t.Errorf("ret is not nil: %#v", ret)
				}
				if err == nil {
					t.Errorf("want error: %v", err)
				}
				w.WriteHeader(http.StatusBadRequest)
			},
			wantErr: true,
			want: &want{
				StatusCode: http.StatusBadRequest,
				Resp:       nil,
			},
		},
	}

	handler := NewGreeterHTTPConverter(&EchoGreeterServer{})

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req, err := tt.reqFunc()
			if err != nil {
				t.Fatal(err)
			}

			rec := httptest.NewRecorder()
			handler.SayHello(tt.cb).ServeHTTP(rec, req)

			var resp *HelloReply
			if !tt.wantErr {
				resp = &HelloReply{}
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
			}

			actual := &want{
				StatusCode: rec.Code,
				Resp:       resp,
			}

			if diff := cmp.Diff(actual, tt.want); diff != "" {
				t.Errorf("%s", diff)
			}
		})
	}
}
