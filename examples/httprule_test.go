package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
)

func TestMessaging_GetMessage(t *testing.T) {
	type want struct {
		StatusCode int
		Method     string
		Path       string
		Resp       *GetMessageResponse
	}
	tests := []struct {
		name    string
		reqFunc func() (*http.Request, error)
		cb      func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error)
		wantErr bool
		want    *want
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
	}

	handler := NewMessagingHTTPConverter(&Messaging{})

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req, err := tt.reqFunc()
			if err != nil {
				t.Fatal(err)
			}

			rec := httptest.NewRecorder()
			method, path, h := handler.GetMessageHTTPRule(tt.cb)
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
					if err := jsonpb.Unmarshal(rec.Body, resp); err != nil {
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

			if diff := cmp.Diff(actual, tt.want); diff != "" {
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
					MessageId: "abc1234",
					Sub: &SubMessage{
						Subfield: "submsg",
					},
					Message: "Hello World!",
				},
			},
		},
	}

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
					if err := jsonpb.Unmarshal(rec.Body, resp); err != nil {
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

			if diff := cmp.Diff(actual, tt.want); diff != "" {
				t.Errorf("%s", diff)
			}
		})
	}
}
