# logger

```
package main

import (
	"log/slog"
	"os"

	"github.com/riete/logger"
)

func main() {
	// create a file writer with DefaultFileRotator
	fw := logger.NewFileWriter("logs/test.log", logger.DefaultFileRotator)
	defer fw.Close() // flush buffered data
	
	// create a network(udp or tcp) writer, 
	nw := logger.NewNetworkWriter("tcp", "127.0.0.1", 10*logger.SizeMiB)
	
	log := logger.New(
		fw,
		logger.WithJSONFormat(),             // use json format
		logger.WithMultiWriter(os.Stdout),   // log to multi target
		logger.WithColor(),                  // use ansi color
		logger.WithLogLevel(slog.LevelInfo), // default log level, lower than it will be ignored
	)
	log.Info("info log")
	log.Warn("warn log")
	log.Error("error log")

}

```