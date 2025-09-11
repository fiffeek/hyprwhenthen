package test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

const (
	binaryPathEnvVar = "HWT_BINARY_PATH"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(filepath.Dir(b))
	binaryPath = ""

	debug      = flag.Bool("debug", false, "use to run the binary in debug mode")
	regenerate = flag.Bool("regenerate", false, "regenerate fixtures instead of comparing")
)

func TestMain(m *testing.M) {
	binaryPath = os.Getenv(binaryPathEnvVar)
	if binaryPath == "" {
		fmt.Printf("no binary provided")
		os.Exit(1)
	}

	os.Exit(m.Run())
}
