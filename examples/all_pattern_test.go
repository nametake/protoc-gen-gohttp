package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
)

func TestAllPattern_AllPattern(t *testing.T) {
	type want struct {
		StatusCode int
		Method     string
		Path       string
		Resp       *AllPatternMessage
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
				req := httptest.NewRequest(http.MethodGet, "/all/pattern", nil)
				req.Header.Set("Content-Type", "application/json")
				return req, nil
			},
			cb: func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error) {
				t.Log(err)
			},
			wantErr: false,
			want: &want{
				StatusCode: http.StatusOK,
				Method:     http.MethodGet,
				Path:       "/all/pattern",
				Resp:       &AllPatternMessage{},
			},
		},
		{
			name: "GET method and Content-Type Protobuf",
			reqFunc: func() (*http.Request, error) {
				req := httptest.NewRequest(http.MethodGet, "/all/pattern", nil)
				req.Header.Set("Content-Type", "application/protobuf")
				return req, nil
			},
			cb:      nil,
			wantErr: false,
			want: &want{
				StatusCode: http.StatusOK,
				Method:     http.MethodGet,
				Path:       "/all/pattern",
				Resp: &AllPatternMessage{
					RepeatedDouble: make([]float64, 0),
				},
			},
		},
	}

	handler := NewAllPatternHTTPConverter(&AllPattern{})

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req, err := tt.reqFunc()
			if err != nil {
				t.Fatal(err)
			}

			rec := httptest.NewRecorder()
			method, path, h := handler.AllPatternHTTPRule(tt.cb)
			h.ServeHTTP(rec, req)

			var resp *AllPatternMessage
			if !tt.wantErr {
				resp = &AllPatternMessage{}
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
