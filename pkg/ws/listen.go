package ws

import (
	"encoding/json"
	"log"
	"net/http"

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
	log.Printf("start server, server listen at [%s]", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
