package main

import (
	"github.com/DenisPavlov/monitoring/internal/logger"
	"github.com/DenisPavlov/monitoring/internal/routing"
	"github.com/DenisPavlov/monitoring/internal/storage"
	"net/http"
)

func main() {
	parseFlags()
	if err := logger.Initialize(flagLogLevel, flagRunEnv); err != nil {
		panic(err)
	}
	memStorage := storage.NewMemStorage()
	router := routing.BuildRouter(memStorage)

	logger.Log.Infoln("Running server on", flagRunAddr)
	if err := http.ListenAndServe(flagRunAddr, router); err != nil {
		panic(err)
	}
}
