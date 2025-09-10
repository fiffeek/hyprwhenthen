package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

var (
	debug      bool
	configPath string
	rootCmd    = &cobra.Command{
		Use:              "hyprwhenthen",
		Short:            "Event-driven automation for Hyprland",
		Long:             "HyprWhenThen is an automation tool that listens to Hyprland events and executes actions based on configured rules.",
		Version:          fmt.Sprintf("%s (commit %s, built %s)", Version, Commit, BuildDate),
		PersistentPreRun: setupLogger,
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func setupLogger(cmd *cobra.Command, args []string) {
	if debug {
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
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVar(
		&configPath,
		"config",
		"$HOME/.config/hyprwhenthen/config.toml",
		"Path to configuration file",
	)
}
