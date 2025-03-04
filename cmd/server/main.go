package main

import (
	"fmt"
	"github.com/DenisPavlov/monitoring/internal/routing"
	"github.com/DenisPavlov/monitoring/internal/storage"
	"net/http"
)

func main() {
	parseFlags()
	memStorage := storage.NewMemStorage()
	router := routing.BuildRouter(memStorage)

	fmt.Println("Running server on", flagRunAddr)
	if err := http.ListenAndServe(flagRunAddr, router); err != nil {
		panic(err)
	}
}
