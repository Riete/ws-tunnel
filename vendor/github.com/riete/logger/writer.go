package logger

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type FileWriter struct {
	f       *os.File
	rotator Rotator
	path    string
	written *atomic.Int64
	mu      sync.Mutex
}

func (f *FileWriter) open() {
	f.f, _ = os.OpenFile(f.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

func (f *FileWriter) loadSize() {
	fs, _ := f.f.Stat()
	f.written.Store(fs.Size())
}

func (f *FileWriter) Close() error {
	return f.f.Close()
}

func (f *FileWriter) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	n, err := f.f.Write(p)
	f.written.Add(int64(n))
	if f.written.Load() >= f.rotator.MaxSize() {
		_ = f.Close()
		f.rotator.Rotate(f.path)
		f.open()
		f.written.Store(0)
	}
	return n, err
}

func NewFileWriter(path string, rotator Rotator) io.WriteCloser {
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	fw := &FileWriter{path: path, rotator: rotator, written: new(atomic.Int64)}
	fw.open()
	fw.loadSize()
	return fw
}

type BufWriter struct {
	w        io.Writer
	bw       *bufio.Writer
	bufSize  int
	interval time.Duration
	mu       sync.Mutex
	stop     chan struct{}
}

func (b *BufWriter) flush() {
	t := time.NewTicker(b.interval)
	defer t.Stop()
	for {
		select {
		case <-b.stop:
			return
		case <-t.C:
			if b.bw.Buffered() > 0 {
				b.mu.Lock()
				_ = b.bw.Flush()
				b.mu.Unlock()
			}
		}
	}
}

func (b *BufWriter) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.bw.Write(p)
}

func (b *BufWriter) Close() error {
	close(b.stop)
	_ = b.bw.Flush()
	if c, ok := b.w.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func NewBufWriter(w io.Writer, options ...BufWriterOption) io.WriteCloser {
	b := &BufWriter{w: w, bufSize: 4096, interval: time.Second, stop: make(chan struct{})}
	for _, option := range options {
		option(b)
	}
	b.bw = bufio.NewWriterSize(w, b.bufSize)
	go b.flush()
	return b
}

type NetworkWriter struct {
	conn       net.Conn
	network    string
	addr       string
	err        error
	maxBufSize int
	buf        *bytes.Buffer
	mu         sync.Mutex
	closed     bool
}

func (n *NetworkWriter) dial() {
	if n.closed {
		n.err = errors.New("writer has been closed")
	}
	n.conn, n.err = net.DialTimeout(n.network, n.addr, 5*time.Second)
}

func (n *NetworkWriter) bufWrite(p []byte) {
	if n.maxBufSize > 0 {
		if n.buf.Len() >= n.maxBufSize {
			// drop old len(p) data
			n.buf.Next(len(p))
		}
		n.buf.Write(p)
	}
}

func (n *NetworkWriter) Write(p []byte) (int, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.err != nil || n.conn == nil {
		n.dial()
		if n.err != nil {
			n.bufWrite(p)
			return 0, n.err
		}
	}

	var rn int
	var bn int
	if n.buf.Len() > 0 {
		if bn, n.err = n.conn.Write(n.buf.Bytes()); n.err != nil {
			n.buf.Next(bn)
			n.bufWrite(p)
			return bn, n.err
		}
		n.buf.Reset()
	}
	if rn, n.err = n.conn.Write(p); n.err != nil {
		n.bufWrite(p)
	}
	return rn + bn, n.err
}

func (n *NetworkWriter) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.closed = true
	if n.err == nil {
		n.err = n.conn.Close()
	}
	return n.err
}

// NewNetworkWriter
// if maxBufSize > 0, when data written to the remote server failed
// it will be cached in memory using a maximum of maxBufSize
// old data will be discarded if the maxBufSize is reached
func NewNetworkWriter(network, addr string, maxBufSize int) io.WriteCloser {
	w := &NetworkWriter{
		network:    network,
		addr:       addr,
		maxBufSize: maxBufSize,
		buf:        new(bytes.Buffer),
	}
	w.dial()
	return w
}
