package test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/fiffeek/hyprwhenthen/internal/hypr"
	"github.com/fiffeek/hyprwhenthen/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func Test__Run_Binary(t *testing.T) {
	tests := []struct {
		name                string
		config              string
		extraArgs           []string
		expectError         bool
		expectErrorContains string
		expectLogsContain   []string
		validateSideEffects func(*testing.T, map[string]string)
		waitForSideEffects  func(context.Context, *testing.T, map[string]string)
		hyprEvents          []string
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
		{
			name:                "should fail when general timeout is missing",
			config:              "testdata/configs/should_fail_missing_general_timeout.toml",
			extraArgs:           []string{"validate"},
			expectError:         true,
			expectErrorContains: "timeout has to be set",
		},
		{
			name:                "should fail when general timeout is negative",
			config:              "testdata/configs/should_fail_negative_general_timeout.toml",
			extraArgs:           []string{"validate"},
			expectError:         true,
			expectErrorContains: "timeout must be positive",
		},
		{
			name:                "should fail when handler timeout is negative",
			config:              "testdata/configs/should_fail_negative_handler_timeout.toml",
			extraArgs:           []string{"validate"},
			expectError:         true,
			expectErrorContains: "timeout must be positive",
		},
		{
			name:                "should fail when handler 'on' field is missing",
			config:              "testdata/configs/should_fail_missing_on_field.toml",
			extraArgs:           []string{"validate"},
			expectError:         true,
			expectErrorContains: "'on' field is required",
		},
		{
			name:                "should fail invalid regex",
			config:              "testdata/configs/should_fail_invalid_regex.toml",
			extraArgs:           []string{"validate"},
			expectError:         true,
			expectErrorContains: "regexp expression is invalid",
		},
		{
			name:                "should fail when handler 'when' field is missing",
			config:              "testdata/configs/should_fail_missing_when_field.toml",
			extraArgs:           []string{"validate"},
			expectError:         true,
			expectErrorContains: "'when' field is required",
		},
		{
			name:                "should fail when handler 'then' field is missing",
			config:              "testdata/configs/should_fail_missing_then_field.toml",
			extraArgs:           []string{"validate"},
			expectError:         true,
			expectErrorContains: "'then' field is required",
		},
		{
			name:        "should fail when no handlers are configured",
			config:      "testdata/configs/should_fail_no_handlers.toml",
			extraArgs:   []string{"validate"},
			expectError: true,
		},
		{
			name:        "should succeed with valid handler timeout",
			config:      "testdata/configs/should_succeed_with_handler_timeout.toml",
			extraArgs:   []string{"validate"},
			expectError: false,
		},
		{
			name:        "should respond to event",
			config:      "testdata/configs/should_respond_to_event.toml",
			extraArgs:   []string{"run"},
			expectError: true,
			hyprEvents: []string{
				"windowtitlev2>>558f74f82570,Mozilla Firefox",
			},
			validateSideEffects: func(t *testing.T, env map[string]string) {
				testutils.AssertFileExists(t, env["TMP_TST_FILE_0"])
			},
			waitForSideEffects: func(ctx context.Context, t *testing.T, env map[string]string) {
				funcs := []func() error{
					func() error {
						return testutils.FileExists(env["TMP_TST_FILE_0"])
					},
				}
				waitTillHolds(ctx, t, funcs, 300*time.Millisecond)
			},
		},
		{
			name:        "should capture groups",
			config:      "testdata/configs/should_capture_groups.toml",
			extraArgs:   []string{"run"},
			expectError: true,
			hyprEvents: []string{
				"windowtitlev2>>558f74f82570,Mozilla Firefox",
			},
			validateSideEffects: func(t *testing.T, env map[string]string) {
				testutils.AssertFileExists(t, env["TMP_TST_FILE_0"])
				compareWithFixture(t, env["TMP_TST_FILE_0"],
					"testdata/fixtures/should_capture_groups")
			},
			waitForSideEffects: func(ctx context.Context, t *testing.T, env map[string]string) {
				funcs := []func() error{
					func() error {
						return testutils.FileExists(env["TMP_TST_FILE_0"])
					},
					func() error {
						return testutils.ContentSameAsFixture(t, env["TMP_TST_FILE_0"],
							"testdata/fixtures/should_capture_groups")
					},
				}
				waitTillHolds(ctx, t, funcs, 300*time.Millisecond)
			},
		},
		{
			name:        "should dispatch regex",
			config:      "testdata/configs/should_dispatch_regex.toml",
			extraArgs:   []string{"run"},
			expectError: true,
			hyprEvents: []string{
				"windowtitlev2>>558f74f82570,Mozilla Firefox",
				"windowtitlev2>>558f74f82571,Google Chrome",
			},
			validateSideEffects: func(t *testing.T, env map[string]string) {
				testutils.AssertFileExists(t, env["TMP_TST_FILE_0"])
				testutils.AssertFileExists(t, env["TMP_TST_FILE_1"])
				compareWithFixture(t, env["TMP_TST_FILE_0"],
					"testdata/fixtures/should_dispatch_regex__0")
				compareWithFixture(t, env["TMP_TST_FILE_1"],
					"testdata/fixtures/should_dispatch_regex__1")
			},
			waitForSideEffects: func(ctx context.Context, t *testing.T, env map[string]string) {
				funcs := []func() error{
					func() error {
						return testutils.ContentSameAsFixture(t, env["TMP_TST_FILE_0"],
							"testdata/fixtures/should_dispatch_regex__0")
					},
					func() error {
						return testutils.ContentSameAsFixture(t, env["TMP_TST_FILE_1"],
							"testdata/fixtures/should_dispatch_regex__1")
					},
				}
				waitTillHolds(ctx, t, funcs, 300*time.Millisecond)
			},
		},
		{
			name:        "should use routing_key",
			config:      "testdata/configs/should_use_routing_key.toml",
			extraArgs:   []string{"run"},
			expectError: true,
			hyprEvents: []string{
				"windowtitlev2>>558f74f82570,Mozilla Firefox",
			},
			validateSideEffects: func(t *testing.T, env map[string]string) {
				testutils.AssertFileExists(t, env["TMP_TST_FILE_0"])
				compareWithFixture(t, env["TMP_TST_FILE_0"],
					"testdata/fixtures/should_use_routing_key")
			},
			waitForSideEffects: func(ctx context.Context, t *testing.T, env map[string]string) {
				funcs := []func() error{
					func() error {
						return testutils.ContentSameAsFixture(t, env["TMP_TST_FILE_0"],
							"testdata/fixtures/should_use_routing_key")
					},
				}
				waitTillHolds(ctx, t, funcs, 300*time.Millisecond)
			},
			expectLogsContain: []string{
				"routing_key=\"558f74f82570\"",
			},
		},
		{
			name:        "should process serially on key",
			config:      "testdata/configs/should_process_serially_on_key.toml",
			extraArgs:   []string{"run"},
			expectError: true,
			hyprEvents: []string{
				"windowtitlev2>>558f74f82570,Mozilla Firefox",
				"windowtitlev2>>558f74f82570,Mozilla Firefox -- Another",
			},
			validateSideEffects: func(t *testing.T, env map[string]string) {
				testutils.AssertFileExists(t, env["TMP_TST_FILE_0"])
				compareWithFixture(t, env["TMP_TST_FILE_0"],
					"testdata/fixtures/should_process_serially_on_key")
			},
			waitForSideEffects: func(ctx context.Context, t *testing.T, env map[string]string) {
				funcs := []func() error{
					func() error {
						return testutils.ContentSameAsFixture(t, env["TMP_TST_FILE_0"],
							"testdata/fixtures/should_process_serially_on_key")
					},
				}
				waitTillHolds(ctx, t, funcs, 400*time.Millisecond)
			},
			expectLogsContain: []string{
				"routing_key=\"558f74f82570\"",
			},
		},

		{
			name:        "should process serially",
			config:      "testdata/configs/should_process_serially.toml",
			extraArgs:   []string{"run", "--workers", "1"},
			expectError: true,
			hyprEvents: []string{
				"windowtitlev2>>558f74f82570,Mozilla Firefox",
				"windowtitlev2>>558f74f82570,Mozilla Firefox -- Another",
			},
			validateSideEffects: func(t *testing.T, env map[string]string) {
				testutils.AssertFileExists(t, env["TMP_TST_FILE_0"])
				compareWithFixture(t, env["TMP_TST_FILE_0"],
					"testdata/fixtures/should_process_serially")
			},
			waitForSideEffects: func(ctx context.Context, t *testing.T, env map[string]string) {
				funcs := []func() error{
					func() error {
						return testutils.ContentSameAsFixture(t, env["TMP_TST_FILE_0"],
							"testdata/fixtures/should_process_serially")
					},
				}
				waitTillHolds(ctx, t, funcs, 400*time.Millisecond)
			},
		},
		{
			name:        "should timeout",
			config:      "testdata/configs/should_timeout.toml",
			extraArgs:   []string{"run", "--workers", "1"},
			expectError: true,
			hyprEvents: []string{
				"windowtitlev2>>558f74f82570,Timeout",
				"windowtitlev2>>558f74f82570,Regular",
			},
			validateSideEffects: func(t *testing.T, env map[string]string) {
				testutils.AssertFileExists(t, env["TMP_TST_FILE_0"])
				compareWithFixture(t, env["TMP_TST_FILE_0"],
					"testdata/fixtures/should_timeout")
			},
			waitForSideEffects: func(ctx context.Context, t *testing.T, env map[string]string) {
				funcs := []func() error{
					func() error {
						return testutils.ContentSameAsFixture(t, env["TMP_TST_FILE_0"],
							"testdata/fixtures/should_timeout")
					},
				}
				waitTillHolds(ctx, t, funcs, 400*time.Millisecond)
			},
		},
		{
			name:        "should route to multiple handlers",
			config:      "testdata/configs/should_route_to_multiple_targets.toml",
			extraArgs:   []string{"run"},
			expectError: true,
			hyprEvents: []string{
				"windowtitlev2>>558f74f82570,Mozilla Firefox",
			},
			validateSideEffects: func(t *testing.T, env map[string]string) {
				testutils.AssertFileExists(t, env["TMP_TST_FILE_0"])
				compareWithFixture(t, env["TMP_TST_FILE_0"],
					"testdata/fixtures/should_route_to_multiple_targets__0")
				testutils.AssertFileExists(t, env["TMP_TST_FILE_1"])
				compareWithFixture(t, env["TMP_TST_FILE_1"],
					"testdata/fixtures/should_route_to_multiple_targets__1")
			},
			waitForSideEffects: func(ctx context.Context, t *testing.T, env map[string]string) {
				funcs := []func() error{
					func() error {
						return testutils.ContentSameAsFixture(t, env["TMP_TST_FILE_0"],
							"testdata/fixtures/should_route_to_multiple_targets__0")
					},
					func() error {
						return testutils.ContentSameAsFixture(t, env["TMP_TST_FILE_1"],
							"testdata/fixtures/should_route_to_multiple_targets__1")
					},
				}
				waitTillHolds(ctx, t, funcs, 400*time.Millisecond)
			},
		},
		{
			name:        "should handle crash recovery",
			config:      "testdata/configs/should_handle_crash_recovery.toml",
			extraArgs:   []string{"run", "--workers", "1"},
			expectError: true,
			hyprEvents: []string{
				"windowtitlev2>>558f74f82570,Crash",
				"windowtitlev2>>558f74f82570,Regular",
			},
			validateSideEffects: func(t *testing.T, env map[string]string) {
				testutils.AssertFileExists(t, env["TMP_TST_FILE_0"])
				compareWithFixture(t, env["TMP_TST_FILE_0"],
					"testdata/fixtures/should_handle_crash_recovery")
			},
			waitForSideEffects: func(ctx context.Context, t *testing.T, env map[string]string) {
				funcs := []func() error{
					func() error {
						return testutils.ContentSameAsFixture(t, env["TMP_TST_FILE_0"],
							"testdata/fixtures/should_handle_crash_recovery")
					},
				}
				waitTillHolds(ctx, t, funcs, 400*time.Millisecond)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binaryStartingChan := make(chan struct{})
			ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
			defer cancel()

			// fake hypr ipc event socket
			var fakeHyprEventServerDone chan struct{}
			if len(tt.hyprEvents) > 0 {
				xdgRuntimeDir, signature := testutils.SetupHyprEnvVars(t)
				eventsListener, teardownEvents := testutils.SetupHyprSocket(ctx, t,
					xdgRuntimeDir, signature, hypr.GetHyprEventsSocket)
				defer teardownEvents()
				fakeHyprEventServerDone = testutils.SetupFakeHyprEventsServer(ctx, t, eventsListener, tt.hyprEvents)
			}

			args := append([]string{
				"--config", tt.config,
			}, tt.extraArgs...)
			if *debug {
				args = append(args, "--debug")
			}

			done := make(chan struct{})
			var out []byte
			var binaryErr error
			tmpDir := t.TempDir()
			extraEnv := prepTestEnv(tmpDir)

			go func() {
				defer close(done)
				cmd := prepBinaryRun(ctx, args, inlineEnv(extraEnv))
				t.Log(cmd.Args)
				close(binaryStartingChan)
				out, binaryErr = cmd.CombinedOutput()
			}()

			if tt.waitForSideEffects != nil {
				testutils.Logf(t, "Starting waitForSideEffects")
				tt.waitForSideEffects(ctx, t, extraEnv)
				testutils.Logf(t, "waitForSideEffects returned, calling cancel()")
				cancel()
			}

			waitFor(t, fakeHyprEventServerDone)

			select {
			case <-time.After(1000 * time.Millisecond):
				assert.True(t, false, "timeout while running, out: %s", string(out))
			case <-done:
				t.Log(string(out))
				if tt.expectError {
					assert.Error(t, binaryErr, "expected run to fail but it succeeded. Output: %s", string(out))
					assert.Contains(t, string(out), tt.expectErrorContains,
						"error message should contain expected substring. Got: %s", string(out))
				} else {
					assert.NoError(t, binaryErr, "expected to exit cleanly")
				}
				for _, expected := range tt.expectLogsContain {
					assert.Contains(t, string(out), expected, "combined logs should contain a substring")
				}
				if tt.validateSideEffects != nil {
					tt.validateSideEffects(t, extraEnv)
				}
			}
		})
	}
}

func inlineEnv(env map[string]string) []string {
	extraEnv := []string{}
	for key, value := range env {
		extraEnv = append(extraEnv, fmt.Sprintf("%s=%s", key, value))
	}
	return extraEnv
}

func prepTestEnv(tmpDir string) map[string]string {
	extraEnv := map[string]string{}
	for i := range 10 {
		file := filepath.Join(tmpDir, fmt.Sprintf("file_%d", i))
		extraEnv[fmt.Sprintf("TMP_TST_FILE_%d", i)] = file
	}
	return extraEnv
}

func waitFor(t *testing.T, server chan struct{}) {
	if server == nil {
		return
	}

	select {
	case <-server:
	case <-time.After(800 * time.Millisecond):
		assert.True(t, false, "Fake server didn't finish in time")
	}
}

func waitTillHolds(ctx context.Context, t *testing.T, funcs []func() error, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	testutils.Logf(t, "waitTillHolds starting, deadline in %v", timeout)

	for {
		select {
		case <-ticker.C:
			allPass := true
			for _, f := range funcs {
				if err := f(); err != nil {
					allPass = false
					break
				}
			}
			if !allPass {
				testutils.Logf(t, "Conditions do not hold yet")
			}
			if allPass {
				testutils.Logf(t, "All conditions hold, returning")
				return
			}
			if time.Now().After(deadline) {
				testutils.Logf(t, "After deadline, exiting")
				return
			}
		case <-ctx.Done():
			testutils.Logf(t, "waitTillHolds: Context cancelled, cause: %v", context.Cause(ctx))
			return
		}
	}
}

func compareWithFixture(t *testing.T, target, fixture string) {
	if *regenerate {
		testutils.UpdateFixture(t, target, fixture)
		return
	}
	testutils.AssertContentsSameAsFixture(t, target, fixture)
}

func prepBinaryRun(ctx context.Context, args, extraEnv []string) *exec.Cmd {
	// nolint:gosec
	cmd := exec.CommandContext(
		ctx,
		filepath.Join(basepath, binaryPath),
		args...)
	cmd.Env = append(os.Environ(), "GOCOVERDIR=.coverdata")
	cmd.Env = append(cmd.Env, extraEnv...)
	return cmd
}
