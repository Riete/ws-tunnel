package cmd

import (
	"log/slog"
	"os"

	"github.com/riete/ws-tunnel/pkg/logger"

	"github.com/riete/ws-tunnel/pkg/ws"

	"github.com/spf13/cobra"
)

var logLevel string

var rootCmd = &cobra.Command{
	Use:   "ws-tunnel",
	Short: "ws tunnel server or client or proxy",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var level slog.Level
		switch logLevel {
		case "debug":
			level = slog.LevelDebug
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		}
		logger.Init(level)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		logger.Close()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&ws.DefaultToke, "token", ws.DefaultToke, "auth token")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level, one of debug, info, warn, error")
}
