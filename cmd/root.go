package cmd

import (
	"os"

	"github.com/riete/ws-tunnel/pkg/ws"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ws-tunnel",
	Short: "ws tunnel server or client or proxy",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&ws.DefaultToke, "token", ws.DefaultToke, "auth token")
}
