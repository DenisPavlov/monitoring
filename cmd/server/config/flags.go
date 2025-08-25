// Package config provides application server configuration functionality.
// It handles parsing of command line flags and environment variables.
package config

import (
	"flag"
	"os"
	"strconv"
)

// Global configuration variables for the server application.
//
// These variables store values obtained from command line flags
// and environment variables. Environment variables take precedence.
var (
	// FlagRunAddr is the address and port to run the server.
	// Format: "host:port". Default: "localhost:8080".
	FlagRunAddr string

	// FlagLogLevel is the logging level.
	// Supported values: "Debug", "Info", "Warn", "Error".
	// Default: "Info".
	FlagLogLevel string

	// FlagRunEnv is the application runtime environment.
	// Examples: "development", "production", "testing".
	// Default: "production".
	FlagRunEnv string

	// FlagStoreInterval is the file storage interval in seconds.
	// Default: 300 seconds (5 minutes).
	FlagStoreInterval int

	// FlagFileStoragePath is the path to the storage file.
	// If empty, data may be stored in memory or not persisted.
	FlagFileStoragePath string

	// FlagRestore indicates whether to restore data from file on startup.
	// If true, data will be loaded from file when application starts.
	// Default: false.
	FlagRestore bool

	// FlagDatabaseDSN is the Data Source Name for database connection.
	// Format depends on the database system used.
	// If empty, file-based or in-memory storage will be used.
	FlagDatabaseDSN string

	// FlagKey is the key used to verify request signatures.
	// Used for security and request authentication purposes.
	FlagKey string
)

// ParseFlags parses command line flags and environment variables.
//
// The function performs:
//  1. Parsing of command line flags with default values
//  2. Reading and applying environment variables (if set)
//  3. Environment variables take precedence over command line flags
//
// Supported environment variables:
//   - ADDRESS: server address and port (equivalent to flag -a)
//   - LOG_LEVEL: logging level (equivalent to flag -l)
//   - RUN_ENV: runtime environment (equivalent to flag -e)
//   - STORE_INTERVAL: storage interval in seconds (equivalent to flag -i)
//   - FILE_STORAGE_PATH: storage file path (equivalent to flag -f)
//   - RESTORE: restore flag (equivalent to flag -r)
//   - DATABASE_DSN: database DSN (equivalent to flag -d)
//   - KEY: signature key (equivalent to flag -k)
//
// Returns an error if:
//   - numeric values (STORE_INTERVAL) cannot be converted
//   - boolean values (RESTORE) cannot be converted
//
// Usage example:
//
//	err := config.ParseFlags()
//	if err != nil {
//	    log.Fatal("Failed to parse flags:", err)
//	}
//
// Command line flags example:
//
//	./app -a localhost:8080 -l Info -i 300 -r true
//
// Environment variables example:
//
//	export ADDRESS=localhost:8080
//	export LOG_LEVEL=Info
//	export STORE_INTERVAL=300
func ParseFlags() error {
	flag.StringVar(&FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&FlagLogLevel, "l", "Info", "log level")
	flag.StringVar(&FlagRunEnv, "e", "production", "Run environment")
	flag.IntVar(&FlagStoreInterval, "i", 300, "File store interval in seconds")
	flag.StringVar(&FlagFileStoragePath, "f", "", "Storage file path")
	flag.BoolVar(&FlagRestore, "r", false, "Load storage data from file")
	flag.StringVar(&FlagDatabaseDSN, "d", "", "Database DSN")
	flag.StringVar(&FlagKey, "k", "", "key used to check the request sign")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		FlagRunAddr = envRunAddr
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		FlagLogLevel = envLogLevel
	}

	if envRunEnv := os.Getenv("RUN_ENV"); envRunEnv != "" {
		FlagRunEnv = envRunEnv
	}

	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		val, err := strconv.Atoi(envStoreInterval)
		if err != nil {
			return err
		}
		FlagStoreInterval = val
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		FlagFileStoragePath = envFileStoragePath
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		val, err := strconv.ParseBool(envRestore)
		if err != nil {
			return err
		}
		FlagRestore = val
	}

	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		FlagDatabaseDSN = envDatabaseDSN
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		FlagKey = envKey
	}

	return nil
}
