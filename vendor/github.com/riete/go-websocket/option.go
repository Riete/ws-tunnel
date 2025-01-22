package ws

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type UpgraderOption func(upgrader *websocket.Upgrader)

func WithHandshakeTimeout(t time.Duration) UpgraderOption {
	return func(upgrader *websocket.Upgrader) {
		upgrader.HandshakeTimeout = t
	}
}

func WithReadBufferSize(b int) UpgraderOption {
	return func(upgrader *websocket.Upgrader) {
		upgrader.ReadBufferSize = b
	}
}

func WithWriteBufferSize(b int) UpgraderOption {
	return func(upgrader *websocket.Upgrader) {
		upgrader.WriteBufferSize = b
	}
}

func WithDisableCheckOrigin() UpgraderOption {
	return func(upgrader *websocket.Upgrader) {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}
}

func WithCheckOrigin(f func(*http.Request) bool) UpgraderOption {
	return func(upgrader *websocket.Upgrader) {
		upgrader.CheckOrigin = f
	}
}

func WithEnableCompression() UpgraderOption {
	return func(upgrader *websocket.Upgrader) {
		upgrader.EnableCompression = true
	}
}
