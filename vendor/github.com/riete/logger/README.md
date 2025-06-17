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
    
    // create a buffer writer
    bw := logger.NewBufWriter(
        fw,                                      // underlying writer, here use previously created "fw"
        logger.WithBufSize(100),                 // set buf size, default is 4096
        logger.WithFlushInterval(time.Second),   // set flush interval, default is 1s
    )
    
    // create a network(udp or tcp) writer
    nw := logger.NewNetworkWriter("tcp", "127.0.0.1", 10*logger.SizeMiB)
    
    log := logger.New(
        bw,
        logger.WithJSONFormat(),                 // use json format
        logger.WithMultiWriter(os.Stdout, nw),   // log to multi target
        logger.WithColor(),                      // use ansi color
        logger.WithLogLevel(slog.LevelInfo),     // set log level, default is info, lower than it will be ignored
        logger.WithDisableCaller(),              // disable log source code position of the log statement, default is enable
        logger.WithCaller("source", 4),          // set key and skip(more than 3 if wrapped), default is "source" and 3 
        logger.WithAttr(slog.Attr{})             // set additional global attr 
    )
    
    defer fw.Close() // close opened file
    defer bw.Close() // flush buffered data, and call underlying writer.Close() if possible
    defer nw.Close() // close network connection
    or
    defer log.Close() // close all io.Closer
    
    log.Debug("debug log")
    log.Info("info log")
    log.Warn("warn log")
    log.Error("error log")
    log.Fatal("fatal log")
    
    log.Debugf("%s", "debug log")
    log.Infof("%s", "info log")
    log.Warnf("%s", "warn log")
    log.Errorf("%s", "error log")
    log.Fatalf("%s", "fatal log")
}

```