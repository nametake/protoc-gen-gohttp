package main

import (
	"context"
	"fmt"
	"net/http"
)

type EchoGreeterServer struct {
}

func (s *EchoGreeterServer) SayHello(ctx context.Context, req *HelloRequest) (*HelloReply, error) {
	return &HelloReply{
		Message: fmt.Sprintf("Hello, %s!", req.Name),
	}, nil
}

func main() {
	srv := &EchoGreeterServer{}

	cnv := NewGreeterHTTPConverter(srv)

	http.Handle("sayhello", cnv.SayHello(nil))

	http.ListenAndServe(":8080", nil)
}
