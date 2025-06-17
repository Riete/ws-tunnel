package logger

import (
	"io"
	"log/slog"
)

type Color string

const (
	// ANSI color, https://gist.github.com/iamnewton/8754917#file-bash-colors-md
	start  Color = "\033["
	end    Color = "\033[0m"
	Black  Color = "30m"
	Red    Color = "31m"
	Green  Color = "32m"
	Yellow Color = "33m"
	Blue   Color = "34m"
	Purple Color = "35m"
	Cyan   Color = "36m"
	White  Color = "37m"
	Gray   Color = "90m"
)

// DefaultColors overwrite or add additional level colors
var DefaultColors = map[slog.Level]Color{
	slog.LevelDebug: Gray,
	slog.LevelWarn:  Yellow,
	slog.LevelError: Red,
}

type colorWriter struct {
	w     io.Writer
	level slog.Level
	lf    string
}

func (c *colorWriter) Write(p []byte) (int, error) {
	color := DefaultColors[c.level]
	if color != "" {
		p = p[0 : len(p)-len(c.lf)]
		p = append([]byte(start+color), append(p, []byte(string(end)+c.lf)...)...)
	}
	return c.w.Write(p)
}
