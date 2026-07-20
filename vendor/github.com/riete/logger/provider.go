package logger

import (
	"log/slog"
	"os"
	"sync/atomic"
)

var loggerProvider = new(atomic.Pointer[Logger])

func SetLogger(l *Logger) {
	loggerProvider.Store(l)
}

func GetLogger() *Logger {
	return loggerProvider.Load()
}

var defaultLogger = New(
	os.Stdout,
	WithColor(),
	WithLogLevel(slog.LevelInfo),
	WithCaller("source", 3),
)

func init() {
	SetLogger(defaultLogger)
}
