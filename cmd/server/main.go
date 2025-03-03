package main

import (
	"github.com/DenisPavlov/monitoring/internal/routing"
	"github.com/DenisPavlov/monitoring/internal/storage"
	"net/http"
)

func main() {
	memStorage := storage.NewMemStorage()
	router := routing.BuildRouter(memStorage)
	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}
