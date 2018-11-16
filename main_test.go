package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGolden(t *testing.T) {
	workdir, err := ioutil.TempDir("", "protoc-gen-gohttp-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(workdir)

	// find all proto file in testdata.
	packages := map[string][]string{}
	if err := filepath.Walk("testdata", func(path string, info os.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".proto") {
			return nil
		}

		dir := filepath.Dir(path)
		packages[dir] = append(packages[dir], path)
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	t.Log(packages)
}
