package config

import (
	"flag"
	"os"
	"strconv"
)

var FlagRunAddr string
var FlagReportInterval int
var FlagPollInterval int
var FlagKey string
var FlagRateLimit int

func ParseFlags() error {
	flag.StringVar(&FlagRunAddr, "a", "localhost:8080", "server address and port")
	flag.IntVar(&FlagReportInterval, "r", 10, "frequency of sending metrics to the server in seconds")
	flag.IntVar(&FlagPollInterval, "p", 2, "frequency of getting runtime metrics in seconds")
	flag.StringVar(&FlagKey, "k", "", "key used to sign the request")
	flag.IntVar(&FlagRateLimit, "l", 5, "rate limit")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		FlagRunAddr = envRunAddr
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		val, err := strconv.Atoi(envReportInterval)
		if err != nil {
			return err
		}
		FlagReportInterval = val
	}
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		val, err := strconv.Atoi(envPollInterval)
		if err != nil {
			return err
		}
		FlagPollInterval = val
	}
	if envKey := os.Getenv("KEY"); envKey != "" {
		FlagKey = envKey
	}
	if envRateLimit := os.Getenv("RATE_LIMIT"); envRateLimit != "" {
		val, err := strconv.Atoi(envRateLimit)
		if err != nil {
			return err
		}
		FlagRateLimit = val
	}

	return nil
}
