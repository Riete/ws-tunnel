package ws

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Conn struct {
	conn           *websocket.Conn
	oriPingHandler func(string) error
}

func (c *Conn) Conn() *websocket.Conn {
	return c.conn
}

func (c *Conn) NetConn() net.Conn {
	return c.conn.NetConn()
}

func (c *Conn) PingHandler() func(string) error {
	return c.conn.PingHandler()
}

func (c *Conn) PongHandler() func(string) error {
	return c.conn.PongHandler()
}

func (c *Conn) CloseHandler() func(int, string) error {
	return c.conn.CloseHandler()
}

func (c *Conn) SetPingHandler(h func(string) error) {
	c.conn.SetPingHandler(func(s string) error {
		if err := h(s); err != nil {
			return err
		}
		return c.oriPingHandler(s)
	})
}

func (c *Conn) SetPongHandler(h func(string) error) {
	c.conn.SetPongHandler(h)
}

func (c *Conn) SetCloseHandler(h func(int, string) error) {
	c.conn.SetCloseHandler(h)
}

func (c *Conn) SetCompressionLevel(level int) error {
	return c.conn.SetCompressionLevel(level)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *Conn) SetReadLimit(limit int64) {
	c.conn.SetReadLimit(limit)
}

func (c *Conn) WritePing(data []byte) error {
	return c.conn.WriteControl(websocket.PingMessage, data, time.Now().Add(time.Second))
}

func (c *Conn) WritePong(data []byte) error {
	return c.conn.WriteControl(websocket.PongMessage, data, time.Now().Add(time.Second))
}

func (c *Conn) WriteClose(code int, text string) error {
	return c.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, text), time.Now().Add(time.Second))
}

func (c *Conn) WriteMessage(data []byte) error {
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *Conn) WriteBinary(data []byte) error {
	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *Conn) WriteJson(v interface{}) error {
	return c.conn.WriteJSON(v)
}

func (c *Conn) ReadMessage() (int, []byte, error) {
	return c.conn.ReadMessage()
}

func (c *Conn) ReadJson(v interface{}) error {
	return c.conn.ReadJSON(v)
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) SendHeartbeat(ctx context.Context, interval time.Duration, threshold int64, data []byte, onQuit func(error)) {
	timeout := time.Duration(threshold) * interval
	go func() {
		var err error
		ticker := time.NewTicker(interval)
		defer func() {
			ticker.Stop()
			if onQuit != nil {
				onQuit(err)
			}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err = c.WritePing(data); err != nil {
					return
				}
			}
		}
	}()
	_ = c.SetReadDeadline(time.Now().Add(timeout))
	customPongHandler := c.PongHandler()
	c.SetPongHandler(func(s string) error {
		_ = c.SetReadDeadline(time.Now().Add(timeout))
		return customPongHandler(s)
	})
}

func NewServer(w http.ResponseWriter, r *http.Request, h http.Header, options ...UpgraderOption) (*Conn, error) {
	upgrader := websocket.Upgrader{}
	for _, option := range options {
		option(&upgrader)
	}
	conn, err := upgrader.Upgrade(w, r, h)
	return &Conn{conn: conn, oriPingHandler: conn.PingHandler()}, err
}

func NewClient(dialer *websocket.Dialer, url string, h http.Header) (*Conn, error) {
	if dialer == nil {
		dialer = websocket.DefaultDialer
	}
	conn, r, err := dialer.Dial(url, h)
	if err == nil {
		return &Conn{conn: conn, oriPingHandler: conn.PingHandler()}, nil
	}
	if r != nil {
		b, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		return nil, fmt.Errorf("connect to [%s] failed: %s, %s", url, err.Error(), string(b))
	}
	return nil, fmt.Errorf("connect to [%s] failed: %s", url, err.Error())
}
