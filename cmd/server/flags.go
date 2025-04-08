package main

import (
	"flag"
	"os"
	"strconv"
)

var (
	flagRunAddr         string
	flagLogLevel        string
	flagRunEnv          string
	flagStoreInterval   int
	flagFileStoragePath string
	flagRestore         bool
	flagDatabaseDSN     string
)

func parseFlags() error {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&flagLogLevel, "l", "Info", "log level")
	flag.StringVar(&flagRunEnv, "e", "production", "Run environment")
	flag.IntVar(&flagStoreInterval, "i", 300, "File store interval in seconds")
	flag.StringVar(&flagFileStoragePath, "f", "", "Storage file path")
	flag.BoolVar(&flagRestore, "r", false, "Load storage data from file")
	flag.StringVar(&flagDatabaseDSN, "d", "", "Database DSN")
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

	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		val, err := strconv.Atoi(envStoreInterval)
		if err != nil {
			return err
		}
		flagStoreInterval = val
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		flagFileStoragePath = envFileStoragePath
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		val, err := strconv.ParseBool(envRestore)
		if err != nil {
			return err
		}
		flagRestore = val
	}

	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		flagDatabaseDSN = envDatabaseDSN
	}

	return nil
}
