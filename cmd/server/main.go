package main

import (
	"fmt"
	storage2 "github.com/DenisPavlov/monitoring/internal/storage"
	"net/http"
	"strconv"
	"strings"
)

const (
	gaugeMetricName   = "gauge"
	counterMetricName = "counter"
	updateBasePath    = "/update/"
)

var storage = storage2.NewMemStorage()

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	http.HandleFunc(updateBasePath, saveMetricsHandler(storage))
	return http.ListenAndServe("localhost:8080", nil)
}

func saveMetricsHandler(storage storage2.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("save metric by path", req.URL.Path)
		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		mType, name, strValue, err := parse(req.URL.Path[len(updateBasePath):])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		switch mType {
		case gaugeMetricName:
			value, err := strconv.ParseFloat(strValue, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			storage.AddGauge(name, value)
		case counterMetricName:
			value, err := strconv.ParseInt(strValue, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			storage.AddCounter(name, value)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
		fmt.Println(storage)
	}
}

func parse(path string) (mType, name, value string, err error) {
	slice := strings.Split(path, "/")
	if len(slice) != 3 {
		return "", "", "", fmt.Errorf("invalid path: %s", path)
	}

	return slice[0], slice[1], slice[2], nil
}
