package main

import (
	"flag"
	"log"
	"os"
	"strconv"
)

var flagRunAddr string
var flagReportInterval int
var flagPollInterval int

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "server address and port")
	flag.IntVar(&flagReportInterval, "r", 10, "frequency of sending metrics to the server in seconds")
	flag.IntVar(&flagPollInterval, "p", 2, "frequency of getting runtime metrics in seconds")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		val, err := strconv.Atoi(envReportInterval)
		if err != nil {
			log.Fatal(err)
		}
		flagReportInterval = val
	}
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		val, err := strconv.Atoi(envPollInterval)
		if err != nil {
			log.Fatal(err)
		}
		flagPollInterval = val
	}
}
