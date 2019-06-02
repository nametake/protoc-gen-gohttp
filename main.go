package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	in, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	var req plugin.CodeGeneratorRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return err
	}

	resp, err := Generate(&req)
	if err != nil {
		return err
	}

	out, err := proto.Marshal(resp)
	if err != nil {
		return err
	}

	if _, err := os.Stdout.Write(out); err != nil {
		return err
	}

	return nil
}
