package logger

import (
	"log/slog"
	"os"

	"github.com/riete/logger"
)

var defaultLogger *logger.Logger

func Init(level slog.Level) {
	fw := logger.NewFileWriter("ws-tunnel.log", logger.NewFileRotator(200*logger.SizeMiB, 1))
	defaultLogger = logger.New(
		fw,
		logger.WithColor(),
		logger.WithLogLevel(level),
		logger.WithMultiWriter(os.Stdout),
		logger.WithCaller("source", 4),
	)
}

func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

func Close() {
	_ = defaultLogger.Close()
}
