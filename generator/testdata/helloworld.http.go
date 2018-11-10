package helloworldpb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

type Greeter struct{}

func NewGreeter() *Greeter {
	return &Greeter{}
}

func (g *Greeter) SayHello(srv GreeterServer, cb func(ctx context.Context,
	w http.ResponseWriter, r *http.Request,
	arg, ret proto.Message, err error),
) http.HandlerFunc {
	if cb == nil {
		cb = func(ctx context.Context,
			w http.ResponseWriter, r *http.Request,
			arg, ret proto.Message, err error,
		) {
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "%v: arg = %v: ret = %v", err, arg, ret)
			}
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			cb(ctx, w, r, nil, nil, err)
			return
		}

		var arg *HelloRequest
		switch r.Header.Get("Content-Type") {
		case "application/protobuf", "application/x-protobuf":
			if err := proto.Unmarshal(body, arg); err != nil {
				cb(ctx, w, r, nil, nil, err)
				return
			}
		default:
			if err := jsonpb.Unmarshal(bytes.NewBuffer(body), arg); err != nil {
				cb(ctx, w, r, nil, nil, err)
				return
			}
		}

		ret, err := srv.SayHello(ctx, arg)
		if err != nil {
			cb(ctx, w, r, arg, ret, err)
			return
		}

		switch r.Header.Get("Content-Type") {
		case "application/protobuf", "application/x-protobuf":
			buf, err := proto.Marshal(ret)
			if err != nil {
				cb(ctx, w, r, arg, ret, err)
				return
			}
			if _, err := io.Copy(w, bytes.NewBuffer(buf)); err != nil {
				cb(ctx, w, r, arg, ret, err)
				return
			}
		default:
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				cb(ctx, w, r, arg, ret, err)
				return
			}
		}

		cb(ctx, w, r, arg, ret, err)
	})
}
