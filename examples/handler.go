package example

import (
	"context"
	"fmt"
)

type HelloWorldServer struct {
}

func (s *HelloWorldServer) SayHello(ctx context.Context, req *HelloRequest) (*HelloReply, error) {
	return &HelloReply{
		Message: fmt.Sprintf("Hello, %s!", req.Name),
	}, nil
}
