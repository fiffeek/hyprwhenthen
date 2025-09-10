package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test__Run_Binary(t *testing.T) {
	tests := []struct {
		name                string
		config              string
		extraArgs           []string
		expectError         bool
		expectErrorContains string
		validateSideEffects func(*testing.T)
	}{
		{
			name:        "should show help",
			extraArgs:   []string{"run", "--help"},
			expectError: false,
		},
		{
			name:                "should fail when config doesnt exist",
			config:              "testdata/configs/should_fail_when_config_doesnt_exist.toml",
			extraArgs:           []string{"validate"},
			expectError:         true,
			expectErrorContains: "not found",
		},
		{
			name:        "should succeed when valid config exists",
			config:      "testdata/configs/should_succeed_when_valid_config_exists.toml",
			extraArgs:   []string{"validate"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binaryStartingChan := make(chan struct{})
			ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
			defer cancel()

			args := append([]string{
				"--config", tt.config,
			}, tt.extraArgs...)
			if *debug {
				args = append(args, "--debug")
			}

			done := make(chan struct{})
			var out []byte
			var binaryErr error

			go func() {
				defer close(done)
				cmd := prepBinaryRun(ctx, args)
				t.Log(cmd.Args)
				close(binaryStartingChan)
				out, binaryErr = cmd.CombinedOutput()
			}()

			select {
			case <-time.After(1000 * time.Millisecond):
				require.NoError(t, ctx.Err(), "timeout while running, out: %s", string(out))
			case <-done:
				t.Log(string(out))
				if tt.expectError {
					require.Error(t, binaryErr, "expected run to fail but it succeeded. Output: %s", string(out))
					require.Contains(t, string(out), tt.expectErrorContains,
						"error message should contain expected substring. Got: %s", string(out))
				} else {
					assert.NoError(t, binaryErr, "expected to exit cleanly")
				}
				if tt.validateSideEffects != nil {
					tt.validateSideEffects(t)
				}
			}
		})
	}
}
