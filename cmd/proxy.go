package cmd

import (
	"log"
	"time"

	"github.com/riete/ws-tunnel/pkg/ws"
	"github.com/spf13/cobra"
)

var proxyListenPort string
var proxyBindIP string

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "start tunnel proxy",
	Run: func(cmd *cobra.Command, args []string) {
		scheme := "ws://"
		if useWss {
			scheme = "wss://"
		}
		for {
			ws.DialAsProxy(scheme+serverAddr+ws.ProxyPath, clientId, proxyBindIP+":"+proxyListenPort)
			log.Println("try re-connect after 5 seconds")
			time.Sleep(5 * time.Second)
		}
	},
}

func init() {
	rootCmd.AddCommand(proxyCmd)
	proxyCmd.Flags().StringVarP(&clientId, "client-id", "i", "test-client", "client id")
	proxyCmd.Flags().BoolVar(&useWss, "wss", false, "if use wss")
	proxyCmd.Flags().StringVarP(&serverAddr, "server-addr", "s", "127.0.0.1:37452", "server addr")
	proxyCmd.Flags().StringVarP(&proxyListenPort, "proxy-listen-port", "p", "2222", "proxy listen port")
	proxyCmd.Flags().StringVarP(&proxyBindIP, "proxy-bind-ip", "b", "127.0.0.1", "proxy bind ip ")
}
