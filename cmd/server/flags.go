package main

import (
	"flag"
	"os"
)

var flagRunAddr string
var flagLogLevel string
var flagRunEnv string

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&flagLogLevel, "l", "Info", "log level")
	flag.StringVar(&flagRunEnv, "e", "production", "Run environment")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		flagLogLevel = envLogLevel
	}

	if envRunEnv := os.Getenv("RUN_ENV"); envRunEnv != "" {
		flagRunEnv = envRunEnv
	}
}
