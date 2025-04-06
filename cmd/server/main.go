package main

import (
	"github.com/DenisPavlov/monitoring/internal/database"
	"github.com/DenisPavlov/monitoring/internal/handler"
	"github.com/DenisPavlov/monitoring/internal/logger"
	"github.com/DenisPavlov/monitoring/internal/storage"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	if err := run(); err != nil {
		logger.Log.Error(err.Error())
	}
}

func run() error {
	err := parseFlags()
	if err != nil {
		return err
	}
	if err = logger.Initialize(flagLogLevel, flagRunEnv); err != nil {
		return err
	}

	db, err := database.InitDB(flagDatabaseDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	store := initStorage()
	router := handler.BuildRouter(store, db)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			logger.Log.Infoln("Server shutting down", sig)
			err = shutDown(flagFileStoragePath, store)
			if err != nil {
				logger.Log.Errorln(err)
			}
			os.Exit(0)
		}
	}()

	go storeMetricsIfNeeded(flagStoreInterval, flagFileStoragePath, store)

	logger.Log.Infoln("Running server on", flagRunAddr)
	if err := http.ListenAndServe(flagRunAddr, router); err != nil {
		return err
	}

	return nil
}

func shutDown(filename string, store storage.Storage) error {
	return storage.SaveToFile(filename, store)
}

func initStorage() (store storage.Storage) {
	if flagDatabaseDSN != "" {
		logger.Log.Infoln("Initializing postgres database storage")
		return storage.NewPostgresStorage()
	}
	if flagFileStoragePath != "" {
		if flagRestore {
			logger.Log.Infoln("Initializing file storage from file", flagFileStoragePath)
			store, err := storage.InitFromFile(flagStoreInterval == 0, flagFileStoragePath)
			if err != nil {
				logger.Log.Error("Can not load storage from file.", err.Error())
				store = storage.NewFileStorage(flagStoreInterval == 0, flagFileStoragePath)
			}
			return store
		} else {
			logger.Log.Infoln("Initializing file storage with new file", flagFileStoragePath)
			return storage.NewFileStorage(flagStoreInterval == 0, flagFileStoragePath)
		}
	}
	logger.Log.Infoln("Initializing memory storage")
	return storage.NewMemStorage()
}

func storeMetricsIfNeeded(flagStoreInterval int, filename string, store storage.Storage) {
	if flagStoreInterval != 0 {
		count := 1
		for {
			if count%flagStoreInterval == 0 {
				logger.Log.Infoln("Store metrics to file ", filename)
				err := storage.SaveToFile(filename, store)
				if err != nil {
					logger.Log.Errorln(err)
				}
			}
			count++
			time.Sleep(time.Second)
		}
	}
}
