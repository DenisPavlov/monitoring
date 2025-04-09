package main

import (
	"database/sql"
	"github.com/DenisPavlov/monitoring/internal/database"
	"github.com/DenisPavlov/monitoring/internal/handler"
	"github.com/DenisPavlov/monitoring/internal/logger"
	"github.com/DenisPavlov/monitoring/internal/storage"
	"net/http"
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

	var store storage.Storage
	store, err = initStorage(db)
	if err != nil {
		logger.Log.Error("Error initializing storage", err)
		store = storage.NewMemStorage()
	}
	router := handler.BuildRouter(store, db)

	//c := make(chan os.Signal, 1)
	//signal.Notify(c, os.Interrupt)
	//go func() {
	//	for sig := range c {
	//		logger.Log.Infoln("Server shutting down", sig)
	//		fileStore, isFileStore := store.(*storage.FileStorage)
	//		if isFileStore {
	//			err := fileStore.SaveToFile()
	//			if err != nil {
	//				logger.Log.Errorln(err)
	//			}
	//		}
	//		os.Exit(0)
	//	}
	//}()

	logger.Log.Infoln("Running server on", flagRunAddr)
	if err := http.ListenAndServe(flagRunAddr, router); err != nil {
		return err
	}

	return nil
}

func initStorage(db *sql.DB) (store storage.Storage, err error) {
	logger.Log.Info("Initializing storage", db) // todo - удалить эту строчку
	//if flagDatabaseDSN != "" {
	//	logger.Log.Infoln("Initializing postgres database storage")
	//	store, err := storage.NewPostgresStorage(context.Background(), db)
	//	if err != nil {
	//		return nil, err
	//	}
	//	return store, nil
	//}
	//if flagFileStoragePath != "" {
	//	var fileStorage *storage.FileStorage
	//	var err error
	//
	//	if flagRestore {
	//		logger.Log.Infoln("Initializing file storage from file", flagFileStoragePath)
	//		fileStorage, err = storage.InitFromFile(flagStoreInterval == 0, flagFileStoragePath)
	//		if err != nil {
	//			logger.Log.Error("Can not load storage from file.", err)
	//			fileStorage = storage.NewFileStorage(flagStoreInterval == 0, flagFileStoragePath)
	//		}
	//	} else {
	//		logger.Log.Infoln("Initializing file storage with new file", flagFileStoragePath)
	//		fileStorage = storage.NewFileStorage(flagStoreInterval == 0, flagFileStoragePath)
	//	}
	//
	//	go storeMetricsIfNeeded(flagStoreInterval, flagFileStoragePath, *fileStorage)
	//	return fileStorage, nil
	//
	//}
	logger.Log.Infoln("Initializing memory storage")
	return storage.NewMemStorage(), nil
}

//func storeMetricsIfNeeded(flagStoreInterval int, filename string, store storage.FileStorage) {
//	if flagStoreInterval != 0 {
//		count := 1
//		for {
//			if count%flagStoreInterval == 0 {
//				logger.Log.Infoln("Store metrics to file ", filename)
//				err := store.SaveToFile()
//				if err != nil {
//					logger.Log.Errorln(err)
//				}
//			}
//			count++
//			time.Sleep(time.Second)
//		}
//	}
//}
