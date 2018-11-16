package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// When the environment variable RUN_AS_PROTOC_GEN_GO is set, we skip running
// tests and instead act as protoc-gen-go. This allows the test binary to
// pass itself to protoc.
func init() {
	if os.Getenv("RUN_AS_PROTOC_GEN_GO") != "" {
		main()
		os.Exit(0)
	}
}

func TestGolden(t *testing.T) {
	workdir, err := ioutil.TempDir("", "protoc-gen-gohttp-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(workdir)

	// Find all the proto files in testdata.
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

	// Compile each package, using this binary as protoc-gen-gohttp.
	for _, sources := range packages {
		args := []string{"--gohttp_out=" + workdir}
		args = append(args, sources...)
		protoc(t, args)
	}

	// Compare each generated file to the golden version.
	if err := filepath.Walk(workdir, func(path string, info os.FileInfo, _ error) error {
		t.Log(path)
		if info.IsDir() {
			return nil
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func protoc(t *testing.T, args []string) {
	cmd := exec.Command("protoc", "--plugin=protoc-gen-gohttp="+os.Args[0])
	cmd.Args = append(cmd.Args, args...)
	// We set the RUN_AS_PROTOC_GEN_GO environment variable to indicate that
	// the subprocess should act as a proto compiler rather than a test.
	cmd.Env = append(os.Environ(), "RUN_AS_PROTOC_GEN_GO=1")
	out, err := cmd.CombinedOutput()
	if len(out) > 0 || err != nil {
		t.Log("RUNNING: ", strings.Join(cmd.Args, " "))
	}
	if len(out) > 0 {
		t.Log(string(out))
	}
	if err != nil {
		t.Fatalf("protoc: %v", err)
	}
}
