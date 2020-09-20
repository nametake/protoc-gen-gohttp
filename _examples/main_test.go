package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type ErrorService struct{}

func (s *ErrorService) SayHello(ctx context.Context, req *HelloRequest) (*HelloReply, error) {
	return nil, errors.New("ERROR")
}

func TestEchoGreeterServer_SayHello(t *testing.T) {
	type want struct {
		StatusCode  int
		ContentType string
		Resp        proto.Message
	}
	tests := []struct {
		name         string
		reqFunc      func() (*http.Request, error)
		service      GreeterHTTPService
		cb           func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error)
		interceptors []grpc.UnaryServerInterceptor
		wantErr      bool
		want         *want
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
				req.Header.Set("Accept", "*/*")
				return req, nil
			},
			service: &EchoGreeterServer{},
			cb:      nil,
			wantErr: false,
			want: &want{
				StatusCode:  200,
				ContentType: "application/json",
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
				req.Header.Set("Accept", "*/*")
				return req, nil
			},
			service: &EchoGreeterServer{},
			cb:      nil,
			wantErr: false,
			want: &want{
				StatusCode:  200,
				ContentType: "application/protobuf",
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
				req.Header.Set("Accept", "*/*")
				return req, nil
			},
			service: &EchoGreeterServer{},
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
				StatusCode:  http.StatusBadRequest,
				ContentType: "",
				Resp:        nil,
			},
		},
		{
			name: "Accept empty",
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
				req.Header.Set("Accept", "")
				return req, nil
			},
			service: &EchoGreeterServer{},
			cb:      nil,
			wantErr: false,
			want: &want{
				StatusCode:  200,
				ContentType: "application/json",
				Resp: &HelloReply{
					Message: "Hello, John!",
				},
			},
		},
		{
			name: "Content-Type JSON, but Accept Protobuf",
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
				req.Header.Set("Accept", "application/protobuf")
				return req, nil
			},
			service: &EchoGreeterServer{},
			cb:      nil,
			wantErr: false,
			want: &want{
				StatusCode:  200,
				ContentType: "application/protobuf",
				Resp: &HelloReply{
					Message: "Hello, John!",
				},
			},
		},
		{
			name: "Content-Type JSON with interceptors",
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
				req.Header.Set("Accept", "*/*")
				return req, nil
			},
			service: &EchoGreeterServer{},
			cb:      nil,
			interceptors: []grpc.UnaryServerInterceptor{
				grpc.UnaryServerInterceptor(
					func(ctx context.Context, arg interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
						ret, err := handler(ctx, arg)
						if err != nil {
							return nil, err
						}
						r := ret.(*HelloReply)
						r.Message = fmt.Sprintf("~%s~", r.Message)
						return r, nil
					},
				),
				grpc.UnaryServerInterceptor(
					func(ctx context.Context, arg interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
						ret, err := handler(ctx, arg)
						if err != nil {
							return nil, err
						}
						r := ret.(*HelloReply)
						r.Message = fmt.Sprintf("\"%s\"", r.Message)
						return r, nil
					},
				),
			},
			wantErr: false,
			want: &want{
				StatusCode:  200,
				ContentType: "application/json",
				Resp: &HelloReply{
					Message: "~\"Hello, John!\"~",
				},
			},
		},
		{
			name: "Content-Type Protobuf with interceptors",
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
				req.Header.Set("Accept", "*/*")
				return req, nil
			},
			service: &EchoGreeterServer{},
			cb:      nil,
			interceptors: []grpc.UnaryServerInterceptor{
				grpc.UnaryServerInterceptor(
					func(ctx context.Context, arg interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
						ret, err := handler(ctx, arg)
						if err != nil {
							return nil, err
						}
						r := ret.(*HelloReply)
						r.Message = fmt.Sprintf("**%s**", r.Message)
						return r, nil
					},
				),
			},
			wantErr: false,
			want: &want{
				StatusCode:  200,
				ContentType: "application/protobuf",
				Resp: &HelloReply{
					Message: "**Hello, Smith!**",
				},
			},
		},
		{
			name: "Default callback error at Content-Type JSON",
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
				req.Header.Set("Accept", "*/*")
				return req, nil
			},
			service: &ErrorService{},
			cb:      nil,
			wantErr: true,
			want: &want{
				StatusCode:  500,
				ContentType: "application/json",
				Resp: &spb.Status{
					Code:    int32(codes.Unknown),
					Message: "ERROR",
				},
			},
		},
		{
			name: "Default callback error at Content-Type Protobuf",
			reqFunc: func() (*http.Request, error) {
				p := &HelloRequest{
					Name: "John",
				}

				buf, err := proto.Marshal(p)
				if err != nil {
					return nil, err
				}

				body := bytes.NewBuffer(buf)
				req := httptest.NewRequest(http.MethodPost, "/", body)
				req.Header.Set("Content-Type", "application/protobuf")
				req.Header.Set("Accept", "*/*")
				return req, nil
			},
			service: &ErrorService{},
			cb:      nil,
			wantErr: true,
			want: &want{
				StatusCode:  500,
				ContentType: "application/protobuf",
				Resp: &spb.Status{
					Code:    int32(codes.Unknown),
					Message: "ERROR",
				},
			},
		},
	}

	opts := cmpopts.IgnoreUnexported(
		HelloRequest{},
		HelloReply{},
		spb.Status{},
	)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req, err := tt.reqFunc()
			if err != nil {
				t.Fatal(err)
			}

			rec := httptest.NewRecorder()
			handler := NewGreeterHTTPConverter(tt.service)
			handler.SayHello(tt.cb, tt.interceptors...).ServeHTTP(rec, req)

			var resp proto.Message
			var contentType string
			if !tt.wantErr {
				resp = &HelloReply{}
				switch contentType = rec.Header().Get("Content-Type"); contentType {
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
			} else if tt.cb == nil {
				// for default callback
				resp = &spb.Status{}
				switch contentType = rec.Header().Get("Content-Type"); contentType {
				case "application/protobuf":
					if err := proto.Unmarshal(rec.Body.Bytes(), resp); err != nil {
						t.Fatal(err)
					}
				case "application/json":
					if err := protojson.Unmarshal(rec.Body.Bytes(), resp); err != nil {
						t.Fatal(err)
					}
				default:
				}
			}

			actual := &want{
				StatusCode:  rec.Code,
				ContentType: contentType,
				Resp:        resp,
			}

			if diff := cmp.Diff(actual, tt.want, opts); diff != "" {
				t.Errorf("%s", diff)
			}
		})
	}
}
