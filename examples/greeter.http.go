package main

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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GreeterHTTPConverter struct {
	srv GreeterServer
}

func NewGreeterHTTPConverter(srv GreeterServer) *GreeterHTTPConverter {
	return &GreeterHTTPConverter{
		srv: srv,
	}
}

func (h *GreeterHTTPConverter) SayHello(cb func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error)) http.HandlerFunc {
}

type GreeterHTTPConverter struct {
	srv GreeterServer
}

func NewGreeterHTTPConverter(srv GreeterServer) *GreeterHTTPConverter {
	return &GreeterHTTPConverter{
		srv: srv,
	}
}

func (h *GreeterHTTPConverter) SayHello(cb func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error)) http.HandlerFunc {
}
