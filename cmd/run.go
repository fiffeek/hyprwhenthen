package cmd

import (
	"context"
	"errors"
	"hyprwhenthen/internal/app"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	workers   int
	queueSize int
	runCmd    = &cobra.Command{
		Use:   "run",
		Short: "Start the HyprWhenThen service",
		Long:  "Start the HyprWhenThen service to listen for Hyprland events and execute configured actions.",
		Run:   run,
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
		"Defines the queue size",
	)
}

func run(cmd *cobra.Command, args []string) {
	logrus.WithField("version", Version).Info("Starting HyprWhenThen")

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(context.Canceled)

	app, err := app.NewApplication(ctx, configPath, workers, queueSize)
	if err != nil {
		logrus.WithError(err).Fatal("Failed on app creation")
	}
	err = app.Run(ctx)
	if errors.Is(err, context.Canceled) {
		logrus.WithError(err).Info("Context cancelled, exiting")
		return
	}
	if err != nil {
		logrus.WithError(err).Fatal("Service failed")
	}
	logrus.Info("Exiting...")
}
