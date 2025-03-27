package routing

import (
	"encoding/json"
	"fmt"
	"github.com/DenisPavlov/monitoring/internal/compress"
	"github.com/DenisPavlov/monitoring/internal/logger"
	"github.com/DenisPavlov/monitoring/internal/models"
	"github.com/DenisPavlov/monitoring/internal/service"
	"github.com/DenisPavlov/monitoring/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

const (
	updateBasePath = "/update"
	getBasePath    = "/value"
)

func BuildRouter(storage storage.Storage) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(compress.GzipMiddleware)
	r.Route(updateBasePath, func(r chi.Router) {
		r.Post("/", updateMetricHandler(storage))
		r.Post("/{mType}/{mName}/{mValue}", saveMetricsHandler(storage))
	})
	r.Route(getBasePath, func(r chi.Router) {
		r.Post("/", getJSONMetricHandler(storage))
		r.Get("/{mType}/{mName}", getMetricHandler(storage))
	})

	r.Get("/", getAllMetricsHandler(storage))
	return r
}

func saveMetricsHandler(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "mType")
		mName := chi.URLParam(r, "mName")
		mValue := chi.URLParam(r, "mValue")

		req, err := metrics.CreateMetrics(mName, mType, mValue)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Log.Errorf("Error creating metrics: %s", err.Error())
			return
		}

		if err = metrics.Save(req, storage); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}

func getMetricHandler(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "mType")
		mName := chi.URLParam(r, "mName")

		res := metrics.Get(mName, mType, storage)
		if res == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if res.Value != nil {
			_, err := w.Write([]byte(fmt.Sprintf("%g", *res.Value)))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
			}
		} else if res.Delta != nil {
			_, err := w.Write([]byte(strconv.FormatInt(*res.Delta, 10)))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
			}
		}
	}
}

func getJSONMetricHandler(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.Metrics

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Log.Error("cannot decode request JSON body", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		res := metrics.Get(req.ID, req.MType, storage)
		if res == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(res); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Log.Error("cannot encode metric JSON body", err)
			return
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

func updateMetricHandler(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.Metrics

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Log.Error("cannot decode request JSON body ", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := metrics.Save(&req, storage); err != nil {
			logger.Log.Error("cannot save request data to storage ", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Log.Error("cannot encode metric JSON body", err)
			return
		}
	}
}
