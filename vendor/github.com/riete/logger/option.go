package logger

import (
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"time"
)

const (
	defaultCallerKey  = "source"
	defaultCallerSkip = 3
)

var defaultCaller = caller{enable: true, key: defaultCallerKey, skip: defaultCallerSkip}

type Option func(*Logger)

func WithJSONFormat() Option {
	return func(l *Logger) {
		l.json = true
	}
}

func WithColor() Option {
	return func(l *Logger) {
		l.color = true
	}
}

func WithLogLevel(level slog.Level) Option {
	return func(l *Logger) {
		l.SetLevel(level)
	}
}

func WithMultiWriter(w io.Writer, others ...io.Writer) Option {
	return func(l *Logger) {
		l.w = io.MultiWriter(append(others, l.w, w)...)
		for _, other := range append(others, w) {
			if closer, ok := other.(io.Closer); ok {
				l.closers = append(l.closers, closer)
			}
		}
	}
}

func WithAttrs(attrs ...slog.Attr) Option {
	return func(l *Logger) {
		l.SetAttrs(attrs...)
	}
}

func WithCaller(key string, skip int) Option {
	return func(l *Logger) {
		l.caller = caller{enable: true, key: key, skip: skip}
	}
}

func WithDisableCaller() Option {
	return func(l *Logger) {
		l.caller = caller{}
	}
}

type caller struct {
	enable bool
	key    string
	skip   int
}

func (c *caller) caller() string {
	_, file, line, _ := runtime.Caller(c.skip)
	return fmt.Sprintf("%s:%d", file, line)
}

type BufWriterOption func(*BufWriter)

func WithBufSize(s int) BufWriterOption {
	return func(b *BufWriter) {
		b.bufSize = s
	}
}

func WithFlushInterval(t time.Duration) BufWriterOption {
	return func(b *BufWriter) {
		b.interval = t
	}
}
