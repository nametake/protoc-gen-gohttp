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
			name: "GET method, Content-Type JSON and Empty body",
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
				Resp: &AllPatternMessage{
					// Empty array
					RepeatedDouble:   make([]float64, 0),
					RepeatedFloat:    make([]float32, 0),
					RepeatedInt32:    make([]int32, 0),
					RepeatedInt64:    make([]int64, 0),
					RepeatedUint32:   make([]uint32, 0),
					RepeatedUint64:   make([]uint64, 0),
					RepeatedFixed32:  make([]uint32, 0),
					RepeatedFixed64:  make([]uint64, 0),
					RepeatedSfixed32: make([]int32, 0),
					RepeatedSfixed64: make([]int64, 0),
					RepeatedBool:     make([]bool, 0),
					RepeatedString:   make([]string, 0),
					RepeatedBytes:    make([][]byte, 0),
				},
			},
		},
		{
			name: "GET method, Content-Type Protobuf and Empty body",
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
				Resp:       &AllPatternMessage{},
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
