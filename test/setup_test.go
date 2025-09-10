package test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
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

	debug = flag.Bool("debug", false, "use to run the binary in debug mode")
)

func prepBinaryRun(ctx context.Context, args []string) *exec.Cmd {
	// nolint:gosec
	cmd := exec.CommandContext(
		ctx,
		filepath.Join(basepath, binaryPath),
		args...)
	cmd.Env = append(os.Environ(), "GOCOVERDIR=.coverdata")
	return cmd
}

func TestMain(m *testing.M) {
	binaryPath = os.Getenv(binaryPathEnvVar)
	if binaryPath == "" {
		fmt.Printf("no binary provided")
		os.Exit(1)
	}

	os.Exit(m.Run())
}
