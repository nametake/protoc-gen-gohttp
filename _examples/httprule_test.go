package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestMessaging_GetMessage(t *testing.T) {
	type want struct {
		StatusCode int
		Method     string
		Path       string
		Resp       *GetMessageResponse
	}
	tests := []struct {
		name         string
		reqFunc      func() (*http.Request, error)
		cb           func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error)
		interceptors []grpc.UnaryServerInterceptor
		wantErr      bool
		want         *want
	}{
		{
			name: "GET method and Content-Type JSON",
			reqFunc: func() (*http.Request, error) {
				req := httptest.NewRequest(http.MethodGet, "/v1/messages/abc1234?message=hello&tags=a&tags=b", nil)
				req.Header.Set("Content-Type", "application/json")
				return req, nil
			},
			cb:      nil,
			wantErr: false,
			want: &want{
				StatusCode: http.StatusOK,
				Method:     http.MethodGet,
				Path:       "/v1/messages/{message_id}",
				Resp: &GetMessageResponse{
					MessageId: "abc1234",
					Message:   "hello",
					Tags:      []string{"a", "b"},
				},
			},
		},
		{
			name: "GET method and Content-Type Protobuf",
			reqFunc: func() (*http.Request, error) {
				req := httptest.NewRequest(http.MethodGet, "/v1/messages/foobar?message=goodbye&tags=one&tags=two", nil)
				req.Header.Set("Content-Type", "application/protobuf")
				return req, nil
			},
			cb:      nil,
			wantErr: false,
			want: &want{
				StatusCode: http.StatusOK,
				Method:     http.MethodGet,
				Path:       "/v1/messages/{message_id}",
				Resp: &GetMessageResponse{
					MessageId: "foobar",
					Message:   "goodbye",
					Tags:      []string{"one", "two"},
				},
			},
		},
		{
			name: "GET method and Content-Type JSON with interceptors",
			reqFunc: func() (*http.Request, error) {
				req := httptest.NewRequest(http.MethodGet, "/v1/messages/abc1234?message=hello&tags=a&tags=b", nil)
				req.Header.Set("Content-Type", "application/json")
				return req, nil
			},
			cb: nil,
			interceptors: []grpc.UnaryServerInterceptor{
				grpc.UnaryServerInterceptor(
					func(ctx context.Context, arg interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
						ret, err := handler(ctx, arg)
						if err != nil {
							return nil, err
						}
						r := ret.(*GetMessageResponse)
						r.Message = fmt.Sprintf("\"%s\"", r.Message)
						return r, nil
					},
				),
			},
			wantErr: false,
			want: &want{
				StatusCode: http.StatusOK,
				Method:     http.MethodGet,
				Path:       "/v1/messages/{message_id}",
				Resp: &GetMessageResponse{
					MessageId: "abc1234",
					Message:   "\"hello\"",
					Tags:      []string{"a", "b"},
				},
			},
		},
		{
			name: "GET method and Content-Type Protobuf",
			reqFunc: func() (*http.Request, error) {
				req := httptest.NewRequest(http.MethodGet, "/v1/messages/foobar?message=goodbye&tags=one&tags=two", nil)
				req.Header.Set("Content-Type", "application/protobuf")
				return req, nil
			},
			cb:      nil,
			wantErr: false,
			interceptors: []grpc.UnaryServerInterceptor{
				grpc.UnaryServerInterceptor(
					func(ctx context.Context, arg interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
						ret, err := handler(ctx, arg)
						if err != nil {
							return nil, err
						}
						r := ret.(*GetMessageResponse)
						r.Message = fmt.Sprintf("**%s**", r.Message)
						return r, nil
					},
				),
			},
			want: &want{
				StatusCode: http.StatusOK,
				Method:     http.MethodGet,
				Path:       "/v1/messages/{message_id}",
				Resp: &GetMessageResponse{
					MessageId: "foobar",
					Message:   "**goodbye**",
					Tags:      []string{"one", "two"},
				},
			},
		},
	}

	opts := cmpopts.IgnoreUnexported(
		GetMessageRequest{},
		GetMessageResponse{},
	)

	handler := NewMessagingHTTPConverter(&Messaging{})

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req, err := tt.reqFunc()
			if err != nil {
				t.Fatal(err)
			}

			rec := httptest.NewRecorder()
			method, path, h := handler.GetMessageHTTPRule(tt.cb, tt.interceptors...)
			h.ServeHTTP(rec, req)

			var resp *GetMessageResponse
			if !tt.wantErr {
				resp = &GetMessageResponse{}
				switch req.Header.Get("Content-Type") {
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
				StatusCode: rec.Code,
				Method:     method,
				Path:       path,
				Resp:       resp,
			}

			if diff := cmp.Diff(actual, tt.want, opts); diff != "" {
				t.Errorf("%s", diff)
			}
		})
	}
}

func TestMessaging_UpdateMessage(t *testing.T) {
	type want struct {
		StatusCode int
		Method     string
		Path       string
		Resp       *UpdateMessageResponse
	}
	tests := []struct {
		name    string
		reqFunc func() (*http.Request, error)
		cb      func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error)
		wantErr bool
		want    *want
	}{
		{
			name: "PUT method and Content-Type JSON",
			reqFunc: func() (*http.Request, error) {
				p := &UpdateMessageRequest{
					Message: "Hello World!",
				}

				body := &bytes.Buffer{}
				if err := json.NewEncoder(body).Encode(p); err != nil {
					return nil, err
				}

				req := httptest.NewRequest(http.MethodPut, "/v1/messages/abc1234/submsg", body)
				req.Header.Set("Content-Type", "application/json")
				return req, nil
			},
			cb:      nil,
			wantErr: false,
			want: &want{
				StatusCode: http.StatusOK,
				Method:     http.MethodPut,
				Path:       "/v1/messages/{message_id}/{sub.subfield}",
				Resp: &UpdateMessageResponse{
					MessageId: "abc1234",
					Sub: &SubMessage{
						Subfield: "submsg",
					},
					Message: "Hello World!",
				},
			},
		},
		{
			name: "PUT method and Content-Type Protobuf",
			reqFunc: func() (*http.Request, error) {
				p := &UpdateMessageRequest{
					Message: "hello world!",
				}

				buf, err := proto.Marshal(p)
				if err != nil {
					return nil, err
				}
				body := bytes.NewBuffer(buf)

				req := httptest.NewRequest(http.MethodPut, "/v1/messages/foobar/sub", body)
				req.Header.Set("Content-Type", "application/protobuf")
				return req, nil
			},
			cb:      nil,
			wantErr: false,
			want: &want{
				StatusCode: http.StatusOK,
				Method:     http.MethodPut,
				Path:       "/v1/messages/{message_id}/{sub.subfield}",
				Resp: &UpdateMessageResponse{
					MessageId: "foobar",
					Sub: &SubMessage{
						Subfield: "sub",
					},
					Message: "hello world!",
				},
			},
		},
	}

	opts := cmpopts.IgnoreUnexported(
		UpdateMessageRequest{},
		UpdateMessageResponse{},
		SubMessage{},
	)

	handler := NewMessagingHTTPConverter(&Messaging{})

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req, err := tt.reqFunc()
			if err != nil {
				t.Fatal(err)
			}

			rec := httptest.NewRecorder()
			method, path, h := handler.UpdateMessageHTTPRule(tt.cb)
			h.ServeHTTP(rec, req)

			var resp *UpdateMessageResponse
			if !tt.wantErr {
				resp = &UpdateMessageResponse{}
				switch req.Header.Get("Content-Type") {
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
				StatusCode: rec.Code,
				Method:     method,
				Path:       path,
				Resp:       resp,
			}

			if diff := cmp.Diff(actual, tt.want, opts); diff != "" {
				t.Errorf("%s", diff)
			}
		})
	}
}

func TestMessaging_CreateMessage(t *testing.T) {
	type want struct {
		StatusCode int
		Method     string
		Path       string
		Resp       *CreateMessageResponse
	}
	tests := []struct {
		name    string
		reqFunc func() (*http.Request, error)
		cb      func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error)
		wantErr bool
		want    *want
	}{
		{
			name: "POST method and Content-Type JSON",
			reqFunc: func() (*http.Request, error) {
				p := &CreateMessageRequest{
					Opt: "option1",
				}

				body := &bytes.Buffer{}
				if err := json.NewEncoder(body).Encode(p); err != nil {
					return nil, err
				}

				req := httptest.NewRequest(http.MethodPost, "/v1/messages/abc1234/subsub/submsg", body)
				req.Header.Set("Content-Type", "application/json")
				return req, nil
			},
			cb:      nil,
			wantErr: false,
			want: &want{
				StatusCode: http.StatusOK,
				Method:     http.MethodPost,
				Path:       "/v1/messages/{message_id}/{msg.sub.subfield}/{sub.subfield}",
				Resp: &CreateMessageResponse{
					MessageId: "abc1234",
					Sub: &SubMessage{
						Subfield: "submsg",
					},
					Msg: &CreateMessageResponse_Message{
						Sub: &SubMessage{
							Subfield: "subsub",
						},
					},
					Opt: "option1",
				},
			},
		},
		{
			name: "POST method and Content-Type Protobuf",
			reqFunc: func() (*http.Request, error) {
				p := &CreateMessageRequest{
					Opt: "option2",
				}

				buf, err := proto.Marshal(p)
				if err != nil {
					return nil, err
				}
				body := bytes.NewBuffer(buf)

				req := httptest.NewRequest(http.MethodPost, "/v1/messages/foobar/Subsub/sub", body)
				req.Header.Set("Content-Type", "application/protobuf")
				return req, nil
			},
			cb:      nil,
			wantErr: false,
			want: &want{
				StatusCode: http.StatusOK,
				Method:     http.MethodPost,
				Path:       "/v1/messages/{message_id}/{msg.sub.subfield}/{sub.subfield}",
				Resp: &CreateMessageResponse{
					MessageId: "foobar",
					Sub: &SubMessage{
						Subfield: "sub",
					},
					Msg: &CreateMessageResponse_Message{
						Sub: &SubMessage{
							Subfield: "Subsub",
						},
					},
					Opt: "option2",
				},
			},
		},
	}

	opts := cmpopts.IgnoreUnexported(
		CreateMessageRequest{},
		CreateMessageResponse{},
		CreateMessageRequest_Message{},
		CreateMessageResponse_Message{},
		SubMessage{},
	)

	handler := NewMessagingHTTPConverter(&Messaging{})

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req, err := tt.reqFunc()
			if err != nil {
				t.Fatal(err)
			}

			rec := httptest.NewRecorder()
			method, path, h := handler.CreateMessageHTTPRule(tt.cb)
			h.ServeHTTP(rec, req)

			var resp *CreateMessageResponse
			if !tt.wantErr {
				resp = &CreateMessageResponse{}
				switch req.Header.Get("Content-Type") {
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
				StatusCode: rec.Code,
				Method:     method,
				Path:       path,
				Resp:       resp,
			}

			if diff := cmp.Diff(actual, tt.want, opts); diff != "" {
				t.Errorf("%s", diff)
			}
		})
	}
}
