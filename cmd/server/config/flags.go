package config

import (
	"flag"
	"os"
	"strconv"
)

var (
	FlagRunAddr         string
	FlagLogLevel        string
	FlagRunEnv          string
	FlagStoreInterval   int
	FlagFileStoragePath string
	FlagRestore         bool
	FlagDatabaseDSN     string
	FlagKey             string
)

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
