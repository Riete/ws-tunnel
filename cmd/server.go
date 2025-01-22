package cmd

import (
	"github.com/riete/ws-tunnel/pkg/ws"
	"github.com/spf13/cobra"
)

var listenPort string
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start tunnel server",
	Run: func(cmd *cobra.Command, args []string) {
		ws.Listen(":" + listenPort)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVarP(&listenPort, "listen-port", "l", "37452", "listen port")
}
