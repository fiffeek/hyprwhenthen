package cmd

import (
	"hyprwhenthen/internal/config"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long:  "Validate the syntax and structure of the HyprWhenThen configuration file.",
	Run:   validate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func validate(cmd *cobra.Command, args []string) {
	logrus.WithField("version", Version).Debug("Validating configuration")
	_, err := config.NewConfig(configPath)
	if err != nil {
		logrus.WithError(err).Fatal("Configuration is invalid")
	}
	logrus.Info("Configuration is valid")
}
