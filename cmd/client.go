package cmd

import (
	"time"

	"github.com/riete/ws-tunnel/pkg/logger"

	"github.com/riete/ws-tunnel/pkg/ws"
	"github.com/spf13/cobra"
)

var clientId string
var useWss bool
var serverAddr string

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "start tunnel client",
	Run: func(cmd *cobra.Command, args []string) {
		scheme := "ws://"
		if useWss {
			scheme = "wss://"
		}
		for {
			ws.DialAsClient(scheme+serverAddr+ws.ClientPath, clientId)
			logger.Warn("try re-connect after 5 seconds")
			time.Sleep(5 * time.Second)
		}
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().StringVarP(&clientId, "client-id", "i", "test-client", "client id")
	clientCmd.Flags().BoolVar(&useWss, "wss", false, "if use wss")
	clientCmd.Flags().StringVarP(&serverAddr, "server-addr", "s", "127.0.0.1:37452", "server addr")
}
