package tunnel

import (
	"context"
	"sync"

	"github.com/riete/ws2ssh"
)

type Tunnel struct {
	T *ws2ssh.SSHTunnel
	C context.Context
}

type TunnelPool struct {
	rw sync.RWMutex
	p  map[string]*Tunnel
}

func (tp *TunnelPool) Get(clientId string) (*Tunnel, bool) {
	tp.rw.RLock()
	defer tp.rw.RUnlock()
	t, ok := tp.p[clientId]
	return t, ok
}

func (tp *TunnelPool) Set(ctx context.Context, clientId string, tunnel *ws2ssh.SSHTunnel) {
	tp.rw.Lock()
	defer tp.rw.Unlock()
	tp.p[clientId] = &Tunnel{T: tunnel, C: ctx}
}

func (tp *TunnelPool) Delete(clientId string) {
	tp.rw.Lock()
	defer tp.rw.Unlock()
	delete(tp.p, clientId)
}

func (tp *TunnelPool) List() []string {
	tp.rw.RLock()
	defer tp.rw.RUnlock()
	var tl []string
	for i := range tp.p {
		tl = append(tl, i)
	}
	if len(tl) == 0 {
		return []string{}
	}
	return tl
}

var defaultTP = &TunnelPool{p: make(map[string]*Tunnel)}

func Get(clientId string) (*Tunnel, bool) {
	return defaultTP.Get(clientId)
}

func Set(ctx context.Context, clientId string, tunnel *ws2ssh.SSHTunnel) {
	defaultTP.Set(ctx, clientId, tunnel)
}

func Delete(clientId string) {
	defaultTP.Delete(clientId)
}

func List() []string {
	return defaultTP.List()
}
