package ws

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/riete/ws-tunnel/pkg/tunnel"
	"github.com/riete/ws2ssh"

	websocket "github.com/riete/go-websocket"
)

const ClientIdKey = "client-id"

var DefaultToke = "ws-tunnel-token"

func ServerForClient(w http.ResponseWriter, r *http.Request) {
	clientId := r.Header.Get(ClientIdKey)
	log.Printf("receive connection request from client [%s]", clientId)
	if _, exist := tunnel.Get(clientId); exist {
		log.Printf("client [%s] already connected, close connection request", clientId)
		return
	}

	connectedAt := r.Header.Get(ConnectedAtKey)
	conn, err := websocket.NewServer(w, r, nil, websocket.WithDisableCheckOrigin())
	if err != nil {
		log.Printf("websocket server setup failed: %s", err.Error())
		return
	}
	defer conn.Close()
	conn.SetPongHandler(func(s string) error {
		log.Printf("receive pong reply from client: [%s], %s, connected-at: [%s]", clientId, s, connectedAt)
		return nil
	})
	conn.SetPingHandler(func(s string) error {
		log.Printf("receive ping from client: [%s], %s, connected-at: [%s]", clientId, s, connectedAt)
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go conn.SendHeartbeat(ctx, 20*time.Second, 3, []byte(fmt.Sprintf("ping sent to client [%s]", clientId)), nil)

	t := ws2ssh.NewSSHTunnel(conn.Conn())
	if err = t.AsClientSide(ws2ssh.NewClientConfig("ws-tunnel", DefaultToke, nil)); err != nil {
		log.Printf("build tunnel client side for client [%s] failed: %s", clientId, err.Error())
		return
	}
	tunnel.Set(ctx, clientId, t)
	defer tunnel.Delete(clientId)

	log.Printf("connection from client [%s] established success", clientId)
	log.Printf("connection from client [%s] disconnected: %v", clientId, t.Wait())
}

func ServerForProxy(w http.ResponseWriter, r *http.Request) {
	clientId := r.Header.Get(ClientIdKey)
	log.Printf("receive connection request from proxy to use client [%s] to setup proxy", clientId)

	connectedAt := r.Header.Get(ConnectedAtKey)
	conn, err := websocket.NewServer(w, r, nil, websocket.WithDisableCheckOrigin())
	if err != nil {
		log.Printf("websocket server setup failed: %s", err.Error())
		return
	}
	defer conn.Close()
	conn.SetPongHandler(func(s string) error {
		log.Printf("receive pong reply from proxy: %s, connected-at: [%s]", s, connectedAt)
		return nil
	})
	conn.SetPingHandler(func(s string) error {
		log.Printf("receive ping from proxy, %s, connected-at: [%s]", s, connectedAt)
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go conn.SendHeartbeat(ctx, 20*time.Second, 3, []byte("ping sent to proxy"), nil)

	t := ws2ssh.NewSSHTunnel(conn.Conn())
	if err = t.AsServerSide(ws2ssh.NewServerConfig("ws-tunnel", DefaultToke, nil)); err != nil {
		log.Printf("build tunnel server side for proxy failed: %s", err.Error())
		return
	}
	log.Printf("connection from proxy to use client [%s] to setup proxy established success", clientId)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				next, exist := tunnel.Get(clientId)
				if !exist {
					log.Printf("[%s] is not connected, waiting for it to connect", clientId)
					time.Sleep(5 * time.Second)
				} else {
					_ = t.HandleOutgoingContext(next.C, ws2ssh.Next(next.T))
				}
			}
		}
	}()
	log.Printf("connection from proxy to use client [%s] to setup proxy disconnected: %v", clientId, t.Wait())
}
