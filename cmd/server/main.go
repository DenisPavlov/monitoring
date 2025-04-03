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

	var memStorage *storage.MemStorage
	if flagRestore {
		store, err := storage.LoadFromFile(flagStoreInterval == 0, flagFileStoragePath)
		if err != nil {
			logger.Log.Error("Can not load storage from file. ", err.Error())
			store = storage.NewMemStorage(flagStoreInterval == 0, flagFileStoragePath)
		}
		memStorage = store
	} else {
		memStorage = storage.NewMemStorage(flagStoreInterval == 0, flagFileStoragePath)
	}

	db, err := database.InitDB(flagDatabaseDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	router := handler.BuildRouter(memStorage, db)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			logger.Log.Infoln("Server shutting down", sig)
			err = shutDown(flagFileStoragePath, memStorage)
			if err != nil {
				logger.Log.Errorln(err)
			}
			os.Exit(0)
		}
	}()

	go storeMetricsIfNeeded(flagStoreInterval, flagFileStoragePath, memStorage)

	logger.Log.Infoln("Running server on", flagRunAddr)
	if err := http.ListenAndServe(flagRunAddr, router); err != nil {
		return err
	}

	return nil
}

func shutDown(filename string, store *storage.MemStorage) error {
	return storage.SaveToFile(filename, store)
}

func storeMetricsIfNeeded(flagStoreInterval int, filename string, store *storage.MemStorage) {
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
