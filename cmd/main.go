package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"hyprwhenthen/internal/app"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func main() {
	var (
		configPath = flag.String("config", "$HOME/.config/hyprwhenthen/config.toml", "Path to configuration file")
		debug      = flag.Bool("debug", false, "Enable debug logging")
		workers    = flag.Int("workers", 2, "Number of background workers")
		queueSize  = flag.Int("queue", 10, "Defines the queue size")
	)
	flag.Parse()
	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: false,
		DisableColors:    false,
		FullTimestamp:    true,
		ForceQuote:       true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			fn := filepath.Base(f.Function)
			file := fmt.Sprintf("%s:%d", filepath.Base(f.File), f.Line)
			return fn, file
		},
	})

	logrus.WithField("version", Version).Debug("Starting HyprWhenThen")

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(context.Canceled)

	app, err := app.NewApplication(ctx, configPath, *workers, *queueSize)
	if err != nil {
		logrus.WithError(err).Fatal("Failed on app creation")
	}
	err = app.Run(ctx)
	if err == nil {
		logrus.Info("Exiting...")
		return
	}
	if errors.Is(err, context.Canceled) {
		logrus.WithError(err).Info("Context cancelled, exiting")
		return
	}

	// otherwise there is a real error
	logrus.WithError(err).Fatal("Service failed")
}
