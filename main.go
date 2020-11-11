package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"os"
	"tic-tac-toe/game"
)

var (
	addr        = flag.String("addr", "0.0.0.0:443", "TCP address to listen to")
	cert        = flag.String("cert", "ssl/cert.pem", "path to tls-cert file")
	key         = flag.String("key", "ssl/key.pem", "path to tls-key file")
	storagePath = flag.String("storagePath", "storage", "path to storage with game files")
	debug       = flag.Bool("debug", false, "print debug messages")
)

func initLogger() *log.Logger {
	flag.Parse()
	logger := log.New()
	logger.SetFormatter(&log.TextFormatter{
		//DisableColors: true,
		TimestampFormat: "060102 15:04:05",
		DisableSorting:  false,
		FullTimestamp:   true,
	})
	logger.SetOutput(os.Stdout)
	if *debug {
		logger.SetLevel(log.DebugLevel)
		logger.Debugln("debug is set")
	} else {
		logger.SetLevel(log.InfoLevel)
	}
	return logger
}

func main() {
	// init logger
	logger := initLogger()
	storage, err := game.NewStorage(*storagePath, logger)
	if err != nil {
		logger.Fatal("can't open game files storage: ", err)
	}

	ws := NewServer(*addr, *cert, *key, storage, logger)

	err = ws.Run()
	if err != nil {
		logger.Fatal("can't start server: ", err)
	}
}
