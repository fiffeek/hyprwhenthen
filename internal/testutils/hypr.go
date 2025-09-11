package testutils

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func SetupHyprEnvVars(t *testing.T) (string, string) {
	tempDir := t.TempDir()
	signature := "test_signature"
	hyprDir := filepath.Join(tempDir, "hypr", signature)
	//nolint:gosec
	if err := os.MkdirAll(hyprDir, 0o755); err != nil {
		t.Fatalf("Failed to create hypr directory: %v", err)
	}

	originalXDG := os.Getenv("XDG_RUNTIME_DIR")
	originalSig := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	t.Cleanup(func() {
		_ = os.Setenv("XDG_RUNTIME_DIR", originalXDG)
		_ = os.Setenv("HYPRLAND_INSTANCE_SIGNATURE", originalSig)
	})

	_ = os.Setenv("XDG_RUNTIME_DIR", tempDir)
	_ = os.Setenv("HYPRLAND_INSTANCE_SIGNATURE", signature)
	return tempDir, signature
}

func SetupHyprSocket(ctx context.Context, t *testing.T, xdgRuntimeDir, signature string,
	hyprSocketFun func(string, string) string,
) (net.Listener, func()) {
	socketPath := hyprSocketFun(xdgRuntimeDir, signature)
	lc := &net.ListenConfig{}
	listener, err := lc.Listen(ctx, "unix", socketPath)
	require.NoError(t, err, "failed to create a test socket %s", socketPath)
	return listener, func() { _ = listener.Close() }
}

func SetupFakeHyprEventsServer(ctx context.Context, t *testing.T, listener net.Listener, events []string) chan struct{} {
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		conn, err := listener.Accept()
		if err != nil {
			t.Errorf("Failed to accept connection: %v", err)
			return
		}

		t.Log("Accepted connection on events socket")

		for _, event := range events {
			if _, err := conn.Write([]byte(event + "\n")); err != nil {
				t.Errorf("Failed to write event: %v", err)
				return
			}
			t.Log("Wrote event on the event socket")
			time.Sleep(10 * time.Millisecond)
		}

		<-ctx.Done()
	}()
	return serverDone
}
