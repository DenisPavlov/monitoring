// Package config provides agent configuration functionality.
// It handles parsing of command line flags and environment variables
// for a metrics collection and reporting agent.
package config

import (
	"flag"
	"os"
	"strconv"
)

// Global configuration variables for the agent application.
//
// These variables store values obtained from command line flags
// and environment variables. Environment variables take precedence.
var (
	// FlagRunAddr is the server address and port to send metrics to.
	// Format: "host:port". Default: "localhost:8080".
	FlagRunAddr string

	// FlagReportInterval is the frequency of sending metrics to the server in seconds.
	// Default: 10 seconds.
	FlagReportInterval int

	// FlagPollInterval is the frequency of collecting runtime metrics in seconds.
	// Default: 2 seconds.
	FlagPollInterval int

	// FlagKey is the cryptographic key used to sign requests for authentication.
	// If empty, requests are sent without signing.
	FlagKey string

	// FlagRateLimit is the maximum number of concurrent requests allowed.
	// Used to limit the load on both client and server.
	// Default: 5 concurrent requests.
	FlagRateLimit int
)

// ParseFlags parses command line flags and environment variables for agent configuration.
//
// The function performs:
//  1. Parsing of command line flags with default values
//  2. Reading and applying environment variables (if set)
//  3. Environment variables take precedence over command line flags
//
// Supported command line flags:
//
//	-a: server address and port (default: "localhost:8080")
//	-r: report interval in seconds (default: 10)
//	-p: poll interval in seconds (default: 2)
//	-k: signing key (default: "")
//	-l: rate limit (default: 5)
//
// Supported environment variables:
//   - ADDRESS: server address and port (equivalent to flag -a)
//   - REPORT_INTERVAL: report interval in seconds (equivalent to flag -r)
//   - POLL_INTERVAL: poll interval in seconds (equivalent to flag -p)
//   - KEY: signing key (equivalent to flag -k)
//   - RATE_LIMIT: rate limit (equivalent to flag -l)
//
// Returns an error if:
//   - numeric values (REPORT_INTERVAL, POLL_INTERVAL, RATE_LIMIT) cannot be converted from strings
//
// Usage example:
//
//	err := config.ParseFlags()
//	if err != nil {
//	    log.Fatal("Failed to parse configuration:", err)
//	}
//
// Command line example:
//
//	./agent -a localhost:8080 -r 10 -p 2 -l 5
//
// Environment variables example:
//
//	export ADDRESS=localhost:8080
//	export REPORT_INTERVAL=10
//	export POLL_INTERVAL=2
//	export RATE_LIMIT=5
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
