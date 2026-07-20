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

const (
	LevelTrace  slog.Level = slog.LevelDebug - 4
	LevelNotice slog.Level = slog.LevelInfo + 2
)

type Logger struct {
	json     bool
	color    bool
	logger   *slog.Logger
	level    *slog.LevelVar
	closers  []io.Closer
	w        io.Writer
	mu       sync.Mutex
	caller   caller
	traceKey any
}

func (l *Logger) SetLevel(level slog.Level) {
	l.level.Set(level)
}

func (l *Logger) SetAttrs(attrs ...slog.Attr) {
	for _, attr := range attrs {
		l.logger = l.logger.With(attr)
	}
}

func (l *Logger) log(level slog.Level, msg string, args ...any) {
	if l.color {
		if cw, ok := l.w.(*colorWriter); ok {
			l.mu.Lock()
			defer l.mu.Unlock()
			cw.level = level
		}
	}
	if l.caller.enable {
		args = append(args, l.caller.key, l.caller.caller())
	}
	l.logger.Log(context.Background(), level, msg, args...)
}

func (l *Logger) Log(level slog.Level, msg string, args ...any) {
	l.log(level, msg, args...)
}

func (l *Logger) Logf(level slog.Level, format string, v ...any) {
	l.log(level, fmt.Sprintf(format, v...))
}

func (l *Logger) Trace(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.log(level, msg, append([]any{"trace_id", ctx.Value(l.traceKey)}, args...)...)
}

func (l *Logger) Tracef(ctx context.Context, level slog.Level, format string, v ...any) {
	l.log(level, fmt.Sprintf(format, v...), "trace_id", ctx.Value(l.traceKey))
}

func (l *Logger) Debug(msg string, args ...any) {
	l.log(slog.LevelDebug, msg, args...)
}

func (l *Logger) Debugf(format string, v ...any) {
	l.log(slog.LevelDebug, fmt.Sprintf(format, v...))
}

func (l *Logger) Info(msg string, args ...any) {
	l.log(slog.LevelInfo, msg, args...)
}

func (l *Logger) Infof(format string, v ...any) {
	l.log(slog.LevelInfo, fmt.Sprintf(format, v...))
}

func (l *Logger) Notice(msg string, args ...any) {
	l.log(LevelNotice, msg, args...)
}

func (l *Logger) Noticef(format string, v ...any) {
	l.log(LevelNotice, fmt.Sprintf(format, v...))
}

func (l *Logger) Warn(msg string, args ...any) {
	l.log(slog.LevelWarn, msg, args...)
}

func (l *Logger) Warnf(format string, v ...any) {
	l.log(slog.LevelWarn, fmt.Sprintf(format, v...))
}

func (l *Logger) Error(msg string, args ...any) {
	l.log(slog.LevelError, msg, args...)
}

func (l *Logger) Errorf(format string, v ...any) {
	l.log(slog.LevelError, fmt.Sprintf(format, v...))
}

func (l *Logger) Fatal(msg string, args ...any) {
	l.log(slog.LevelError, msg, args...)
	_ = l.Close()
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.log(slog.LevelError, fmt.Sprintf(format, v...))
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
	l.log(slog.LevelInfo, msg, v[1:]...)
}

func (l *Logger) Println(v ...any) {
	if len(v) == 0 {
		v = append(v, "")
	}
	msg, ok := v[0].(string)
	if !ok {
		msg = fmt.Sprintf("%v", v[0])
	}
	l.log(slog.LevelInfo, msg, v[1:]...)
}

func (l *Logger) Printf(format string, v ...any) {
	l.log(slog.LevelInfo, fmt.Sprintf(format, v...))
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
	l := &Logger{level: new(slog.LevelVar), w: w, caller: defaultCaller, traceKey: "trace_id"}
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
	levelReplace := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.LevelKey {
			level, ok := a.Value.Any().(slog.Level)
			if !ok {
				return a
			}
			switch level {
			case LevelTrace:
				return slog.String(a.Key, "TRACE")
			case LevelNotice:
				return slog.String(a.Key, "NOTICE")
			}
			return a
		}
		return a
	}

	var handler slog.Handler
	if l.json {
		handler = slog.NewJSONHandler(l.w, &slog.HandlerOptions{Level: l.level, ReplaceAttr: levelReplace})
	} else {
		handler = slog.NewTextHandler(l.w, &slog.HandlerOptions{Level: l.level, ReplaceAttr: levelReplace})
	}
	l.logger = slog.New(handler)
	return l
}
