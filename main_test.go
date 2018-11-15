package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestGolden(t *testing.T) {
	workdir, err := ioutil.TempDir("", "protoc-gen-gohttp-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(workdir)
}
