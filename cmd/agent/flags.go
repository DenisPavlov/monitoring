package main

import (
	"flag"
	"os"
	"strconv"
)

var flagRunAddr string
var flagReportInterval int
var flagPollInterval int
var flagKey string
var flagRateLimit int

func parseFlags() error {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "server address and port")
	flag.IntVar(&flagReportInterval, "r", 10, "frequency of sending metrics to the server in seconds")
	flag.IntVar(&flagPollInterval, "p", 2, "frequency of getting runtime metrics in seconds")
	flag.StringVar(&flagKey, "k", "", "key used to sign the request")
	flag.IntVar(&flagRateLimit, "l", 5, "rate limit")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		val, err := strconv.Atoi(envReportInterval)
		if err != nil {
			return err
		}
		flagReportInterval = val
	}
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		val, err := strconv.Atoi(envPollInterval)
		if err != nil {
			return err
		}
		flagPollInterval = val
	}
	if envKey := os.Getenv("KEY"); envKey != "" {
		flagKey = envKey
	}
	if envRateLimit := os.Getenv("RATE_LIMIT"); envRateLimit != "" {
		val, err := strconv.Atoi(envRateLimit)
		if err != nil {
			return err
		}
		flagRateLimit = val
	}

	return nil
}
