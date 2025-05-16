package logger

import (
	"io"
	"log/slog"
	"os"

	"github.com/riete/logger"
)

var fw io.WriteCloser
var defaultLogger *logger.Logger

func Init(level slog.Level) {
	rotator := logger.NewFileRotator(200*logger.SizeMiB, 1)
	fw = logger.NewFileWriter("ws-tunnel.log", rotator)
	options := []logger.Option{logger.WithColor(), logger.WithLogLevel(level)}
	if level == slog.LevelDebug {
		options = append(options, logger.WithMultiWriter(os.Stdout))
	}
	defaultLogger = logger.New(fw, options...)
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
	_ = fw.Close()
}
