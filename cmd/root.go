package cmd

import (
	"github.com/djcass44/go-utils/logging"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var command = &cobra.Command{
	Use:          "container-build-engine",
	Short:        "build images",
	SilenceUsage: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logLevel, _ := cmd.Flags().GetInt(flagLogLevel)

		zc := zap.NewProductionConfig()
		zc.Level = zap.NewAtomicLevelAt(zapcore.Level(logLevel * -1))

		_, ctx := logging.NewZap(cmd.Context(), zc)
		cmd.SetContext(ctx)
	},
}

const flagLogLevel = "v"

func init() {
	command.PersistentFlags().Int(flagLogLevel, 0, "log level. Higher is more")
	command.AddCommand(buildCmd)
}

func Execute(version string) {
	command.Version = version
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
