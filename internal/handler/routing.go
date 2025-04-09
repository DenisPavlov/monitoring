package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/DenisPavlov/monitoring/internal/logger"
	"github.com/DenisPavlov/monitoring/internal/models"
	"github.com/DenisPavlov/monitoring/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

const (
	updateBasePath = "/update"
	getBasePath    = "/value"
)

// todo -попробовать выпилить db
func BuildRouter(storage storage.Storage, db *sql.DB) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(GzipMiddleware)
	r.Route(updateBasePath, func(r chi.Router) {
		r.Post("/", updateMetricHandler(storage))
		r.Post("/{mType}/{mName}/{mValue}", saveMetricsHandler(storage))
	})
	r.Route(getBasePath, func(r chi.Router) {
		r.Post("/", getJSONMetricHandler(storage))
		r.Get("/{mType}/{mName}", getMetricHandler(storage))
	})
	r.Get("/ping", pingDBHandler(db))
	r.Post("/updates", func(writer http.ResponseWriter, request *http.Request) {

	})
	r.Get("/", getAllMetricsHandler(storage))
	return r
}

func pingDBHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			logger.Log.Error("Error pinging database: ", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}
}

func saveMetricsHandler(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "mType")
		mName := chi.URLParam(r, "mName")
		mValue := chi.URLParam(r, "mValue")

		req, err := models.CreateMetrics(mName, mType, mValue)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Log.Errorf("Error creating metrics: %s", err.Error())
			return
		}

		if err = storage.Save(r.Context(), req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}

func getMetricHandler(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "mType")
		mName := chi.URLParam(r, "mName")

		res, err := storage.GetByTypeAndID(r.Context(), mName, mType)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if (reflect.DeepEqual(res, models.Metrics{})) {
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

		res, err := storage.GetByTypeAndID(r.Context(), req.ID, req.MType)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if (reflect.DeepEqual(res, models.Metrics{})) {
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

		gauges, err := storage.GetAllByType(r.Context(), models.GaugeMetricName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		for _, value := range gauges {
			page += fmt.Sprintf("\n<p>%s - %f</p>", value.ID, *value.Value)
		}

		counters, err := storage.GetAllByType(r.Context(), models.CounterMetricName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		for _, value := range counters {
			page += fmt.Sprintf("\n<p>%s - %d</p>", value.ID, *value.Delta)
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

		if err := storage.Save(r.Context(), &req); err != nil {
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
