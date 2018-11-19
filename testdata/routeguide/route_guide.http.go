package routeguidepb

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

type RouteGuideHandler struct {
	srv RouteGuideServer
}

func NewRouteGuideHandler(srv RouteGuideServer) *RouteGuideHandler {
	return &RouteGuideHandler{
		srv: srv,
	}
}

func (h *RouteGuideHandler) GetFeature(cb func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error)) http.HandlerFunc {
	if cb == nil {
		cb = func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error) {
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "%v: arg = %v: ret = %v", err, arg, ret)
			}
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx = r.Context()
			arg *Point
			ret *Feature
			err error
		)
		defer cb(ctx, w, r, arg, ret, err)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return
		}

		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/protobuf", "application/x-protobuf":
			if err = proto.Unmarshal(body, arg); err != nil {
				return
			}
		case "application/json":
			if err = jsonpb.Unmarshal(bytes.NewBuffer(body), arg); err != nil {
				return
			}
		default:
			w.WriteHeader(http.StatusUnsupportedMediaType)
			_, err = fmt.Fprintf(w, "Unsupported Content-Type: %s", contentType)
			return
		}

		ret, err = h.srv.GetFeature(ctx, arg)
		if err != nil {
			return
		}

		switch contentType {
		case "application/protobuf", "application/x-protobuf":
			buf, err := proto.Marshal(ret)
			if err != nil {
				return
			}
			if _, err = io.Copy(w, bytes.NewBuffer(buf)); err != nil {
				return
			}
		case "application/json":
			if err = json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
		default:
			w.WriteHeader(http.StatusUnsupportedMediaType)
			_, err = fmt.Fprintf(w, "Unsupported Content-Type: %s", contentType)
			return
		}
	})
}
