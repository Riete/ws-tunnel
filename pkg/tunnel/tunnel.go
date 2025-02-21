package tunnel

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/riete/ws2ssh"
)

type Tunnel struct {
	T            *ws2ssh.SSHTunnel
	C            context.Context
	proxyAddr    string
	proxyStarted bool
	l            sync.Mutex
}

func (t *Tunnel) ProxyServerAddr() string {
	t.l.Lock()
	defer t.l.Unlock()
	return t.proxyAddr
}

func (t *Tunnel) ProxyServerStarted() bool {
	t.l.Lock()
	defer t.l.Unlock()
	return t.proxyStarted
}

func (t *Tunnel) StartProxyServer(proxyAddr string) error {
	t.l.Lock()
	defer t.l.Unlock()
	if t.proxyStarted {
		return nil
	}
	proxyServer := t.T.BuildSocks5ProxyServer()
	proxyStartErr := make(chan error)
	go func() {
		if err := proxyServer.ListenAndServeContext(t.C, proxyAddr); err != nil {
			proxyStartErr <- fmt.Errorf("proxy server fail to listen at [%s]: %s", proxyAddr, err.Error())
		}
	}()
	select {
	case err := <-proxyStartErr:
		log.Println(err.Error())
		return err
	case <-time.After(3 * time.Second):
		log.Printf("start proxy server success, proxy server listen at [%s]", proxyAddr)
	}
	t.proxyAddr = proxyAddr
	t.proxyStarted = true
	return nil
}

type TunnelPool struct {
	rw sync.RWMutex
	t  map[string]*Tunnel
}

func (tp *TunnelPool) Get(clientId string) (*Tunnel, bool) {
	tp.rw.RLock()
	defer tp.rw.RUnlock()
	t, ok := tp.t[clientId]
	return t, ok
}

func (tp *TunnelPool) Set(ctx context.Context, clientId string, tunnel *ws2ssh.SSHTunnel) *Tunnel {
	tp.rw.Lock()
	defer tp.rw.Unlock()
	t := &Tunnel{T: tunnel, C: ctx}
	tp.t[clientId] = t
	return t
}

func (tp *TunnelPool) Delete(clientId string) {
	tp.rw.Lock()
	defer tp.rw.Unlock()
	delete(tp.t, clientId)
}

func (tp *TunnelPool) List() []string {
	tp.rw.RLock()
	defer tp.rw.RUnlock()
	var tl []string
	for i := range tp.t {
		tl = append(tl, i)
	}
	if len(tl) == 0 {
		return []string{}
	}
	return tl
}

var defaultTP = &TunnelPool{t: make(map[string]*Tunnel)}

func Get(clientId string) (*Tunnel, bool) {
	return defaultTP.Get(clientId)
}

func Set(ctx context.Context, clientId string, tunnel *ws2ssh.SSHTunnel) *Tunnel {
	return defaultTP.Set(ctx, clientId, tunnel)
}

func Delete(clientId string) {
	defaultTP.Delete(clientId)
}

func List() []string {
	return defaultTP.List()
}
