package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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
					String_:          "",
					Bytes:            nil,
					RepeatedDouble:   nil,
					RepeatedFloat:    nil,
					RepeatedInt32:    nil,
					RepeatedInt64:    nil,
					RepeatedUint32:   nil,
					RepeatedUint64:   nil,
					RepeatedFixed32:  nil,
					RepeatedFixed64:  nil,
					RepeatedSfixed32: nil,
					RepeatedSfixed64: nil,
					RepeatedBool:     nil,
					RepeatedString:   nil,
					RepeatedBytes:    nil,
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
		{
			name: "GET method, Content-Type JSON",
			reqFunc: func() (*http.Request, error) {
				req := httptest.NewRequest(
					http.MethodGet,
					"/all/pattern?"+
						"double=1.1"+
						"&float=2.2"+
						"&int32=3"+
						"&int64=4"+
						"&uint32=5"+
						"&uint64=6"+
						"&fixed32=7"+
						"&fixed64=8"+
						"&sfixed32=9"+
						"&sfixed64=10"+
						"&bool=true"+
						"&string=hello"+
						"&bytes=aGk="+
						"&repeated_double=11.1"+
						"&repeated_double=11.2"+
						"&repeated_float=12.1"+
						"&repeated_float=12.2"+
						"&repeated_int32=13"+
						"&repeated_int32=14"+
						"&repeated_int64=15"+
						"&repeated_int64=16"+
						"&repeated_uint32=17"+
						"&repeated_uint32=18"+
						"&repeated_uint64=19"+
						"&repeated_uint64=20"+
						"&repeated_fixed32=21"+
						"&repeated_fixed32=22"+
						"&repeated_fixed64=23"+
						"&repeated_fixed64=24"+
						"&repeated_sfixed32=25"+
						"&repeated_sfixed32=26"+
						"&repeated_sfixed64=27"+
						"&repeated_sfixed64=28"+
						"&repeated_bool=false"+
						"&repeated_bool=true"+
						"&repeated_string=hello"+
						"&repeated_string=good%20bye"+
						"&repeated_bytes=ZG9n"+
						"&repeated_bytes=Y2F0",
					nil,
				)
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
					Double:           1.1,
					Float:            2.2,
					Int32:            3,
					Int64:            4,
					Uint32:           5,
					Uint64:           6,
					Fixed32:          7,
					Fixed64:          8,
					Sfixed32:         9,
					Sfixed64:         10,
					Bool:             true,
					String_:          "hello",
					Bytes:            []byte("hi"),
					RepeatedDouble:   []float64{11.1, 11.2},
					RepeatedFloat:    []float32{12.1, 12.2},
					RepeatedInt32:    []int32{13, 14},
					RepeatedInt64:    []int64{15, 16},
					RepeatedUint32:   []uint32{17, 18},
					RepeatedUint64:   []uint64{19, 20},
					RepeatedFixed32:  []uint32{21, 22},
					RepeatedFixed64:  []uint64{23, 24},
					RepeatedSfixed32: []int32{25, 26},
					RepeatedSfixed64: []int64{27, 28},
					RepeatedBool:     []bool{false, true},
					RepeatedString:   []string{"hello", "good bye"},
					RepeatedBytes:    [][]byte{[]byte("dog"), []byte("cat")},
				},
			},
		},
		{
			name: "GET method, Content-Type Protobuf",
			reqFunc: func() (*http.Request, error) {
				req := httptest.NewRequest(
					http.MethodGet,
					"/all/pattern?"+
						"double=1.1"+
						"&float=2.2"+
						"&int32=3"+
						"&int64=4"+
						"&uint32=5"+
						"&uint64=6"+
						"&fixed32=7"+
						"&fixed64=8"+
						"&sfixed32=9"+
						"&sfixed64=10"+
						"&bool=true"+
						"&string=hello"+
						"&bytes=aGk="+
						"&repeated_double=11.1"+
						"&repeated_double=11.2"+
						"&repeated_float=12.1"+
						"&repeated_float=12.2"+
						"&repeated_int32=13"+
						"&repeated_int32=14"+
						"&repeated_int64=15"+
						"&repeated_int64=16"+
						"&repeated_uint32=17"+
						"&repeated_uint32=18"+
						"&repeated_uint64=19"+
						"&repeated_uint64=20"+
						"&repeated_fixed32=21"+
						"&repeated_fixed32=22"+
						"&repeated_fixed64=23"+
						"&repeated_fixed64=24"+
						"&repeated_sfixed32=25"+
						"&repeated_sfixed32=26"+
						"&repeated_sfixed64=27"+
						"&repeated_sfixed64=28"+
						"&repeated_bool=false"+
						"&repeated_bool=true"+
						"&repeated_string=hello"+
						"&repeated_string=good%20bye"+
						"&repeated_bytes=ZG9n"+
						"&repeated_bytes=Y2F0",
					nil,
				)
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
					Double:           1.1,
					Float:            2.2,
					Int32:            3,
					Int64:            4,
					Uint32:           5,
					Uint64:           6,
					Fixed32:          7,
					Fixed64:          8,
					Sfixed32:         9,
					Sfixed64:         10,
					Bool:             true,
					String_:          "hello",
					Bytes:            []byte("hi"),
					RepeatedDouble:   []float64{11.1, 11.2},
					RepeatedFloat:    []float32{12.1, 12.2},
					RepeatedInt32:    []int32{13, 14},
					RepeatedInt64:    []int64{15, 16},
					RepeatedUint32:   []uint32{17, 18},
					RepeatedUint64:   []uint64{19, 20},
					RepeatedFixed32:  []uint32{21, 22},
					RepeatedFixed64:  []uint64{23, 24},
					RepeatedSfixed32: []int32{25, 26},
					RepeatedSfixed64: []int64{27, 28},
					RepeatedBool:     []bool{false, true},
					RepeatedString:   []string{"hello", "good bye"},
					RepeatedBytes:    [][]byte{[]byte("dog"), []byte("cat")},
				},
			},
		},
	}

	opts := cmpopts.IgnoreUnexported(AllPatternMessage{})

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
