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
		Method string
		Path   string
		Resp   *GetMessageResponse
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
				req := httptest.NewRequest(http.MethodGet, `/v1/messages/abc1234?message=hello&tags=a&tags=b`, nil)
				req.Header.Set("Content-Type", "application/json")
				return req, nil
			},
			cb:      nil,
			wantErr: false,
			want: &want{
				Method: http.MethodGet,
				Path:   "/v1/messages/{message_id}",
				Resp: &GetMessageResponse{
					MessageId: "abc1234",
					Message:   "hello",
					Tags:      []string{"a", "b"},
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

			// Check in callback
			if tt.wantErr {
				return
			}

			resp := &GetMessageResponse{}
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
			if diff := cmp.Diff(&want{Method: method, Path: path, Resp: resp}, tt.want); diff != "" {
				t.Errorf("%s", diff)
			}
		})
	}
}

func TestMessaging_UpdateMessage(t *testing.T) {
	type want struct {
		Method string
		Path   string
		Resp   *UpdateMessageResponse
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
				p := &UpdateMessageRequest{
					MessageId: "abc1234",
					Sub: &SubMessage{
						Subfield: "submsg",
					},
					Message: "Hello World!",
				}

				body := &bytes.Buffer{}
				if err := json.NewEncoder(body).Encode(p); err != nil {
					return nil, err
				}

				req := httptest.NewRequest(http.MethodPut, `/v1/messages/abc1234/submsg`, body)
				req.Header.Set("Content-Type", "application/json")
				return req, nil
			},
			cb:      nil,
			wantErr: false,
			want: &want{
				Method: http.MethodPut,
				Path:   "/v1/messages/{message_id}/{sub.subfield}",
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

			// Check in callback
			if tt.wantErr {
				return
			}

			resp := &UpdateMessageResponse{}
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
			if diff := cmp.Diff(&want{Method: method, Path: path, Resp: resp}, tt.want); diff != "" {
				t.Errorf("%s", diff)
			}
		})
	}
}
