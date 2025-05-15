package ws

import (
	"encoding/json"
	"net/http"

	"github.com/riete/ws-tunnel/pkg/logger"

	"github.com/riete/ws-tunnel/pkg/tunnel"
)

const (
	ProxyPath  = "/proxy"
	ClientPath = "/client"
)

func TunnelList(w http.ResponseWriter, r *http.Request) {
	tl := tunnel.List()
	b, _ := json.Marshal(tl)
	_, _ = w.Write(b)
}

func Listen(addr string) {
	http.HandleFunc(ProxyPath, ServerForProxy)
	http.HandleFunc(ClientPath, ServerForClient)
	http.HandleFunc("/tunnel-list", TunnelList)
	logger.Info("server started", "listen_at", addr)
	logger.Error(http.ListenAndServe(addr, nil).Error())
}
