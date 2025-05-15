package ws

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/riete/ws-tunnel/pkg/logger"

	"github.com/riete/ws-tunnel/pkg/tunnel"
	"github.com/riete/ws2ssh"

	websocket "github.com/riete/go-websocket"
)

const ClientIdKey = "client-id"

var DefaultToke = "ws-tunnel-token"

func ServerForClient(w http.ResponseWriter, r *http.Request) {
	clientId := r.Header.Get(ClientIdKey)
	logger.Info("receive connection request from client", "client_id", clientId)
	if _, exist := tunnel.Get(clientId); exist {
		logger.Warn("client already connected, close connection request", "client_id", clientId)
		return
	}

	connectedAt := r.Header.Get(ConnectedAtKey)
	conn, err := websocket.NewServer(w, r, nil, websocket.WithDisableCheckOrigin())
	if err != nil {
		logger.Error(fmt.Sprintf("websocket server setup failed: %s", err.Error()))
		return
	}
	defer conn.Close()
	conn.SetPongHandler(func(s string) error {
		logger.Debug(fmt.Sprintf("receive pong reply from client: %s", s), "client_id", clientId, "connected_at", connectedAt)
		return nil
	})
	conn.SetPingHandler(func(s string) error {
		logger.Debug(fmt.Sprintf("receive ping from client: %s", s), "client_id", clientId, "connected_at", connectedAt)
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go conn.SendHeartbeat(ctx, 20*time.Second, 3, []byte(fmt.Sprintf("ping sent to client [%s]", clientId)), nil)

	t := ws2ssh.NewSSHTunnel(conn.Conn())
	if err = t.AsClientSide(ws2ssh.NewClientConfig("ws-tunnel", DefaultToke, nil)); err != nil {
		logger.Error(fmt.Sprintf("build tunnel client side for client failed: %s", err.Error()), "client_id", clientId)
		return
	}
	tunnel.Set(ctx, clientId, t)
	defer tunnel.Delete(clientId)

	logger.Info("connection from client established success", "client_id", clientId)
	logger.Warn(fmt.Sprintf("connection from client disconnected: %v", t.Wait()), "client_id", clientId)
}

func ServerForProxy(w http.ResponseWriter, r *http.Request) {
	clientId := r.Header.Get(ClientIdKey)
	logger.Info("receive connection request from proxy to use client to setup proxy", "client_id", clientId)

	connectedAt := r.Header.Get(ConnectedAtKey)
	conn, err := websocket.NewServer(w, r, nil, websocket.WithDisableCheckOrigin())
	if err != nil {
		logger.Error(fmt.Sprintf("websocket server setup failed: %s", err.Error()))
		return
	}
	defer conn.Close()
	conn.SetPongHandler(func(s string) error {
		logger.Debug(fmt.Sprintf("receive pong reply from proxy: %s", s), "connected_at", connectedAt)
		return nil
	})
	conn.SetPingHandler(func(s string) error {
		logger.Debug(fmt.Sprintf("receive ping from proxy, %s", s), "connected_at", connectedAt)
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go conn.SendHeartbeat(ctx, 20*time.Second, 3, []byte("ping sent to proxy"), nil)

	t := ws2ssh.NewSSHTunnel(conn.Conn())
	if err = t.AsServerSide(ws2ssh.NewServerConfig("ws-tunnel", DefaultToke, nil)); err != nil {
		logger.Error(fmt.Sprintf("build tunnel server side for proxy failed: %s", err.Error()))
		return
	}
	logger.Info("connection from proxy to use client to setup proxy established success", "client_id", clientId)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				next, exist := tunnel.Get(clientId)
				if !exist {
					logger.Warn("client is not connected, waiting for it to connect", "client_id", clientId)
					time.Sleep(5 * time.Second)
				} else {
					_ = t.HandleOutgoingContext(next.C, ws2ssh.Next(next.T))
				}
			}
		}
	}()
	logger.Warn(fmt.Sprintf("connection from proxy to use client to setup proxy disconnected: %v", t.Wait()), "client_id", clientId)
}
