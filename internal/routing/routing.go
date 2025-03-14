package routing

import (
	"fmt"
	"github.com/DenisPavlov/monitoring/internal/logger"
	"github.com/DenisPavlov/monitoring/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

const (
	gaugeMetricName   = "gauge"
	counterMetricName = "counter"
	updateBasePath    = "/update/"
	getBasePath       = "/value/"
)

func BuildRouter(storage storage.Storage) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Route(updateBasePath+"{mType}/{mName}/{mValue}", func(r chi.Router) {
		r.Post("/", saveMetricsHandler(storage))
	})
	r.Get(getBasePath+"{mType}/{mName}", getMetricHandler(storage))
	r.Get("/", getAllMetricsHandler(storage))
	return r
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

func getMetricHandler(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "mType")
		mName := chi.URLParam(r, "mName")

		switch mType {
		case gaugeMetricName:
			metricValue, ok := storage.Gauge(mName)
			if ok {
				_, err := w.Write([]byte(fmt.Sprintf("%g", metricValue)))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(err.Error()))
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		case counterMetricName:
			metricValue, ok := storage.Counter(mName)
			if ok {
				_, err := w.Write([]byte(strconv.FormatInt(metricValue, 10)))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(err.Error()))
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}
	}
}

func getAllMetricsHandler(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := "<!DOCTYPE html><html><body>"

		gauges := storage.AllGauges()
		for key, value := range gauges {
			page += fmt.Sprintf("\n<p>%s - %f</p>", key, value)
		}

		counters := storage.AllCounters()
		for key, value := range counters {
			page += fmt.Sprintf("\n<p>%s - %d</p>", key, value)
		}

		page += "\n</body></html>"

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page))
	}
}
