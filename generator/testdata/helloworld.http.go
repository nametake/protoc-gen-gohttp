package helloworldpb

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gogo/protobuf/jsonpb"
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
				fmt.Fprintf(w, "%v", err)
			}
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var arg *HelloRequest

		buf, err := ioutil.ReadAll(r.Body)
		if err != nil {
			cb(ctx, w, r, nil, nil, err)
			return
		}

		switch r.Header.Get("Content-Type") {
		case "application/protobuf", "application/x-protobuf":
			if err := proto.Unmarshal(buf, arg); err != nil {
				cb(ctx, w, r, nil, nil, err)
				return
			}
		default:
			if err := jsonpb.Unmarshal(bytes.NewBuffer(buf), arg); err != nil {
				cb(ctx, w, r, nil, nil, err)
				return
			}
		}

		ret, err := srv.SayHello(ctx, arg)
		if err != nil {
			cb(ctx, w, r, arg, ret, err)
			return
		}

		cb(ctx, w, r, arg, ret, err)
	})
}
