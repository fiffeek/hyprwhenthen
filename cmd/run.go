package cmd

import (
	"context"
	"fmt"

	"github.com/fiffeek/hyprwhenthen/internal/app"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	workers   int
	queueSize int
	runCmd    = &cobra.Command{
		Use:           "run",
		Short:         "Start the HyprWhenThen service",
		Long:          "Start the HyprWhenThen service to listen for Hyprland events and execute configured actions.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          run,
	}
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().IntVar(
		&workers,
		"workers",
		2,
		"Number of background workers",
	)
	runCmd.Flags().IntVar(
		&queueSize,
		"queue",
		10,
		"Events are queued for each worker, this defines the queue size; the dispatcher will wait for a free slot when the worker is running behind",
	)
}

func run(cmd *cobra.Command, args []string) error {
	logrus.WithField("version", Version).Info("Starting HyprWhenThen")

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(context.Canceled)

	app, err := app.NewApplication(ctx, cancel, configPath, workers, queueSize)
	if err != nil {
		return fmt.Errorf("failed on app creation: %w", err)
	}
	if err := app.Run(ctx); err != nil {
		return fmt.Errorf("run failed: %w", err)
	}
	return nil
}
