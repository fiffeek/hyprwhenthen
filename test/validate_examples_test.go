package test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/fiffeek/hyprwhenthen/internal/utils"
	"github.com/stretchr/testify/require"
)

var examples = filepath.Join(basepath, "examples")

func Test_Validate_Examples(t *testing.T) {
	files, err := utils.Find(examples, ".toml")
	require.NoError(t, err, "didnt find all example configs")
	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()

			done := make(chan any, 1)
			defer close(done)

			go func() {
				cmd := prepBinaryRun(ctx, []string{"--config", file, "validate"}, []string{})
				out, err := cmd.CombinedOutput()
				require.NoError(t, err, "binary failed %s", string(out))
				done <- true
			}()

			select {
			case <-ctx.Done():
				require.NoError(t, ctx.Err(), "timeout")
			case <-done:
			}
		})
	}
}
