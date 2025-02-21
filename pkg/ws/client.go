package ws

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/riete/ws2ssh"

	websocket "github.com/riete/go-websocket"
)

const ConnectedAtKey = "connected-at"

func header(clientId, connectedAt string) http.Header {
	h := http.Header{}
	h.Set(ConnectedAtKey, connectedAt)
	h.Set(ClientIdKey, clientId)
	return h
}

func dial(url, clientId string) (*websocket.Conn, error) {
	log.Printf("try to connect to server [%s]", url)
	connectedAt := time.Now().Format(time.DateTime)
	h := header(clientId, connectedAt)
	conn, err := websocket.NewClient(nil, url, h)
	if err != nil {
		log.Printf("connect failed: %s", err.Error())
		return conn, err
	}
	conn.SetPongHandler(func(s string) error {
		log.Printf("receive pong reply from server: %s, connected-at: [%s]", s, connectedAt)
		return nil
	})
	conn.SetPingHandler(func(s string) error {
		log.Printf("receive ping from server: %s, connected-at: [%s]", s, connectedAt)
		return nil
	})
	return conn, nil
}

func DialAsClient(url, clientId string) {
	conn, err := dial(url, clientId)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go conn.SendHeartbeat(ctx, 20*time.Second, 3, []byte(fmt.Sprintf("ping sent from client [%s]", clientId)), nil)

	t := ws2ssh.NewSSHTunnel(conn.Conn())
	if err = t.AsServerSide(ws2ssh.NewServerConfig("ws-tunnel", DefaultToke, nil)); err != nil {
		log.Printf("build tunnel server side failed: %s", err.Error())
		return
	}
	log.Println("connection established success")
	go t.HandleOutgoing(ws2ssh.Direct) // nolint: errcheck
	log.Printf("connection lost: %v", t.Wait())
}

func DialAsProxy(url, clientId, proxyAddr string) {
	conn, err := dial(url, clientId)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	oriCloseHandler := conn.CloseHandler()
	conn.SetCloseHandler(func(i int, s string) error {
		cancel()
		return oriCloseHandler(i, s)
	})
	go conn.SendHeartbeat(ctx, 20*time.Second, 3, []byte("ping sent from proxy"), nil)

	t := ws2ssh.NewSSHTunnel(conn.Conn())
	if err = t.AsClientSide(ws2ssh.NewClientConfig("ws-tunnel", DefaultToke, nil)); err != nil {
		log.Printf("build tunnel client side failed: %s", err.Error())
		return
	}
	log.Println("connection established success")

	proxyServer := t.BuildSocks5ProxyServer()
	proxyStartErr := make(chan string)
	go func() {
		if err = proxyServer.ListenAndServeContext(ctx, proxyAddr); err != nil {
			proxyStartErr <- fmt.Sprintf("proxy server fail to listen at [%s]: %s", proxyAddr, err.Error())
		}
	}()
	select {
	case m := <-proxyStartErr:
		log.Println(m)
		return
	case <-time.After(3 * time.Second):
		log.Printf("start proxy server success, proxy server listen at [%s]", proxyAddr)
	}
	log.Printf("connection lost: %v, proxy server [%s] quit", t.Wait(), proxyAddr)
}
