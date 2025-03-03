package routing

import (
	"fmt"
	"github.com/DenisPavlov/monitoring/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

const (
	gaugeMetricName   = "gauge"
	counterMetricName = "counter"
	updateBasePath    = "/update/"
)

func BuildRouter(storage storage.Storage) chi.Router {
	r := chi.NewRouter()
	r.Use(onlyPostMethod)
	r.Route(updateBasePath+"{mType}/{mName}/{mValue}", func(r chi.Router) {
		r.Post("/", saveMetricsHandler(storage))
	})
	return r
}

func onlyPostMethod(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(rw, r)
	})
}

func saveMetricsHandler(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "mType")
		mName := chi.URLParam(r, "mName")
		mValue := chi.URLParam(r, "mValue")

		switch mType {
		case gaugeMetricName:
			value, err := strconv.ParseFloat(mValue, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			storage.AddGauge(mName, value)
		case counterMetricName:
			value, err := strconv.ParseInt(mValue, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			storage.AddCounter(mName, value)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
		fmt.Println(storage)
	}
}
