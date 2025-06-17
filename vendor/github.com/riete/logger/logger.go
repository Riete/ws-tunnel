package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"sync"
)

type Logger struct {
	json    bool
	color   bool
	logger  *slog.Logger
	level   *slog.LevelVar
	closers []io.Closer
	w       io.Writer
	mu      sync.Mutex
	caller  caller
}

func (l *Logger) SetLevel(level slog.Level) {
	l.level.Set(level)
}

func (l *Logger) SetAttrs(attrs ...slog.Attr) {
	for _, attr := range attrs {
		l.logger = l.logger.With(attr)
	}
}

func (l *Logger) Log(level slog.Level, msg string, args ...any) {
	if l.color {
		if cw, ok := l.w.(*colorWriter); ok {
			l.mu.Lock()
			defer l.mu.Unlock()
			cw.level = level
		}
	}
	if l.caller.enable {
		args = append([]any{l.caller.key, l.caller.caller()}, args...)
	}
	l.logger.Log(context.Background(), level, msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.Log(slog.LevelDebug, msg, args...)
}

func (l *Logger) Debugf(format string, v ...any) {
	l.Log(slog.LevelDebug, fmt.Sprintf(format, v...))
}

func (l *Logger) Info(msg string, args ...any) {
	l.Log(slog.LevelInfo, msg, args...)
}

func (l *Logger) Infof(format string, v ...any) {
	l.Log(slog.LevelInfo, fmt.Sprintf(format, v...))
}

func (l *Logger) Warn(msg string, args ...any) {
	l.Log(slog.LevelWarn, msg, args...)
}

func (l *Logger) Warnf(format string, v ...any) {
	l.Log(slog.LevelWarn, fmt.Sprintf(format, v...))
}

func (l *Logger) Error(msg string, args ...any) {
	l.Log(slog.LevelError, msg, args...)
}

func (l *Logger) Errorf(format string, v ...any) {
	l.Log(slog.LevelError, fmt.Sprintf(format, v...))
}

func (l *Logger) Fatal(msg string, args ...any) {
	l.Log(slog.LevelError, msg, args...)
	_ = l.Close()
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.Log(slog.LevelError, fmt.Sprintf(format, v...))
	_ = l.Close()
	os.Exit(1)
}

func (l *Logger) Print(v ...any) {
	if len(v) == 0 {
		v = append(v, "")
	}
	msg, ok := v[0].(string)
	if !ok {
		msg = fmt.Sprintf("%v", v[0])
	}
	l.Log(slog.LevelInfo, msg, v[1:]...)
}

func (l *Logger) Println(v ...any) {
	if len(v) == 0 {
		v = append(v, "")
	}
	msg, ok := v[0].(string)
	if !ok {
		msg = fmt.Sprintf("%v", v[0])
	}
	l.Log(slog.LevelInfo, msg, v[1:]...)
}

func (l *Logger) Printf(format string, v ...any) {
	l.Log(slog.LevelInfo, fmt.Sprintf(format, v...))
}

func (l *Logger) Write(p []byte) (int, error) {
	return l.w.Write(p)
}

func (l *Logger) Close() error {
	for _, c := range l.closers {
		_ = c.Close()
	}
	return nil
}

func New(w io.Writer, options ...Option) *Logger {
	l := &Logger{level: new(slog.LevelVar), w: w, caller: defaultCaller}
	if closer, ok := w.(io.Closer); ok {
		l.closers = append(l.closers, closer)
	}
	for _, option := range options {
		option(l)
	}
	if l.color {
		if runtime.GOOS == "windows" {
			l.w = &colorWriter{w: l.w, lf: "\r\n"}
		} else {
			l.w = &colorWriter{w: l.w, lf: "\n"}
		}
	}
	var handler slog.Handler
	if l.json {
		handler = slog.NewJSONHandler(l.w, &slog.HandlerOptions{Level: l.level})
	} else {
		handler = slog.NewTextHandler(l.w, &slog.HandlerOptions{Level: l.level})
	}
	l.logger = slog.New(handler)
	return l
}
