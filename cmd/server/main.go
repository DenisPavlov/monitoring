package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "net/http/pprof"

	"github.com/DenisPavlov/monitoring/cmd/server/config"
	"github.com/DenisPavlov/monitoring/internal/database"
	"github.com/DenisPavlov/monitoring/internal/handler"
	"github.com/DenisPavlov/monitoring/internal/logger"
	"github.com/DenisPavlov/monitoring/internal/storage"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil {
		logger.Log.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	err := config.ParseFlags()
	if err != nil {
		return err
	}
	if err = logger.Initialize(config.FlagLogLevel, config.FlagRunEnv); err != nil {
		return err
	}

	db, err := database.InitDB(config.FlagDatabaseDSN)
	if err != nil {
		return err
	}
	defer func() {
		_ = db.Close()
	}()

	var store storage.MetricsStorage
	store, err = initStorage(db)
	if err != nil {
		logger.Log.Error("Error initializing storage", err)
		store = storage.NewMemStorage()
	}

	if fileStorage, ok := store.(*storage.FileMetricsStorage); ok {
		go storeMetricsIfNeeded(config.FlagStoreInterval, config.FlagFileStoragePath, fileStorage)
	}

	router := handler.BuildRouter(store, db, config.FlagKey)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			logger.Log.Infoln("Server shutting down", sig)
			fileStore, isFileStore := store.(*storage.FileMetricsStorage)
			if isFileStore {
				err := fileStore.SaveToFile()
				if err != nil {
					logger.Log.Errorln(err)
				}
			}
			os.Exit(0)
		}
	}()

	g, _ := errgroup.WithContext(context.Background())
	g.Go(func() error {
		logger.Log.Infoln("Running server on", config.FlagRunAddr)
		if err := http.ListenAndServe(config.FlagRunAddr, router); err != nil {
			return err
		}
		return nil
	})

	g.Go(func() error {
		logger.Log.Infoln("Running debug server on", "localhost:8082")
		if err := http.ListenAndServe("localhost:8082", nil); err != nil {
			return err
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

func initStorage(db *sql.DB) (store storage.MetricsStorage, err error) {
	if config.FlagDatabaseDSN != "" {
		return initDBStorage(db)
	}
	if config.FlagFileStoragePath != "" {
		return initFileStorage()
	}
	return initMemoryStorage()
}

func initMemoryStorage() (storage.MetricsStorage, error) {
	logger.Log.Infoln("Initializing memory storage")
	return storage.NewMemStorage(), nil
}

func initFileStorage() (storage.MetricsStorage, error) {
	var fileStorage *storage.FileMetricsStorage
	var err error

	if config.FlagRestore {
		logger.Log.Infoln("Initializing file storage from file", config.FlagFileStoragePath)
		fileStorage, err = storage.InitFromFile(config.FlagStoreInterval == 0, config.FlagFileStoragePath)
		if err != nil {
			logger.Log.Error("Can not load storage from file.", err)
			fileStorage = storage.NewFileStorage(config.FlagStoreInterval == 0, config.FlagFileStoragePath)
		}
	} else {
		logger.Log.Infoln("Initializing file storage with new file", config.FlagFileStoragePath)
		fileStorage = storage.NewFileStorage(config.FlagStoreInterval == 0, config.FlagFileStoragePath)
	}

	return fileStorage, nil
}

func initDBStorage(db *sql.DB) (storage.MetricsStorage, error) {
	logger.Log.Infoln("Initializing postgres database storage")
	store, err := storage.NewPostgresStorage(db)
	if err != nil {
		return nil, err
	}
	err = store.InitSchema(context.Background())
	if err != nil {
		return nil, err
	}
	return store, nil
}

func storeMetricsIfNeeded(flagStoreInterval int, filename string, store *storage.FileMetricsStorage) {
	if flagStoreInterval != 0 {
		count := 1
		for {
			if count%flagStoreInterval == 0 {
				logger.Log.Infoln("Store metrics to file ", filename)
				err := store.SaveToFile()
				if err != nil {
					logger.Log.Errorln(err)
				}
			}
			count++
			time.Sleep(time.Second)
		}
	}
}
