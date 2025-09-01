package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/DenisPavlov/monitoring/internal/logger"
	"github.com/DenisPavlov/monitoring/internal/models"
	"github.com/DenisPavlov/monitoring/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	updateBasePath = "/update"
	getBasePath    = "/value"
)

// BuildRouter constructs and configures the chi router with all application routes.
//
// The router includes middleware for:
//   - Request logging
//   - Gzip compression/decompression
//   - SHA256 signature verification (if signKey is provided)
//   - 60-second request timeout
//
// Routes configured:
//   - POST /update/ - Update metric via JSON
//   - POST /update/{mType}/{mName}/{mValue} - Update metric via URL parameters
//   - POST /value/ - Get metric via JSON request
//   - GET /value/{mType}/{mName} - Get metric via URL parameters
//   - GET /ping - Database health check
//   - POST /updates/ - Batch update multiple metrics
//   - GET / - Get all metrics as HTML page
//
// Parameters:
//   - storage: MetricsStorage implementation for data persistence
//   - db: Database connection for health checks
//   - signKey: Cryptographic key for request signature verification (empty disables)
//
// Returns:
//   - chi.Router: Configured router with all middleware and routes
func BuildRouter(storage storage.MetricsStorage, db *sql.DB, signKey string) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(GzipMiddleware)
	if signKey != "" {
		r.Use(SHA256SignMiddleware(signKey))
	}
	r.Use(middleware.Timeout(60 * time.Second))
	r.Route(updateBasePath, func(r chi.Router) {
		r.Post("/", updateMetricHandler(storage))
		r.Post("/{mType}/{mName}/{mValue}", saveMetricsHandler(storage))
	})
	r.Route(getBasePath, func(r chi.Router) {
		r.Post("/", getJSONMetricHandler(storage))
		r.Get("/{mType}/{mName}", getMetricHandler(storage))
	})
	r.Get("/ping", pingDBHandler(db))
	r.Post("/updates/", updatesHandler(storage))
	r.Get("/", getAllMetricsHandler(storage))
	return r
}

// pingDBHandler returns a handler for database health checks.
//
// The handler pings the database with a 1-second timeout and returns:
//   - HTTP 200 if the database is reachable
//   - HTTP 500 if the database connection fails
func pingDBHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		defer func() {
			if err := recover(); r != nil {
				logger.Log.Error("Error pinging database: ", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		if err := db.PingContext(ctx); err != nil {
			logger.Log.Error("Error pinging database: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// saveMetricsHandler returns a handler for saving metrics via URL parameters.
//
// URL format: /update/{mType}/{mName}/{mValue}
//
// Parameters:
//   - mType: Metric type ("gauge" or "counter")
//   - mName: Metric name/identifier
//   - mValue: Metric value (float for gauge, integer for counter)
//
// Returns:
//   - HTTP 400 for invalid parameters or save errors
//   - HTTP 200 on successful save
func saveMetricsHandler(storage storage.MetricsStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "mType")
		mName := chi.URLParam(r, "mName")
		mValue := chi.URLParam(r, "mValue")

		req, err := models.CreateMetric(mName, mType, mValue)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Log.Errorf("Error creating metrics: %s", err.Error())
			return
		}

		if err = storage.Save(r.Context(), req); err != nil {
			logger.Log.Errorf("Error creating metrics: %s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}

// getMetricHandler returns a handler for retrieving metrics via URL parameters.
//
// URL format: /value/{mType}/{mName}
//
// Parameters:
//   - mType: Metric type ("gauge" or "counter")
//   - mName: Metric name/identifier
//
// Returns:
//   - HTTP 400 for invalid parameters or retrieval errors
//   - HTTP 404 if metric not found
//   - HTTP 200 with metric value in response body
func getMetricHandler(storage storage.MetricsStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "mType")
		mName := chi.URLParam(r, "mName")

		res, err := storage.GetByTypeAndID(r.Context(), mName, mType)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if (reflect.DeepEqual(res, models.Metric{})) {
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

// getJSONMetricHandler returns a handler for retrieving metrics via JSON request.
//
// Expected JSON request body format:
//
//	{"id": "metricName", "mType": "gauge|counter"}
//
// Returns:
//   - HTTP 400 for invalid JSON or retrieval errors
//   - HTTP 404 if metric not found
//   - HTTP 200 with metric data in JSON format
func getJSONMetricHandler(storage storage.MetricsStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.Metric

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
		if (reflect.DeepEqual(res, models.Metric{})) {
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

// updatesHandler returns a handler for batch updating multiple metrics.
//
// Expected JSON request body format: array of Metric objects
//
//	[{"id": "name1", "mType": "gauge", "value": 1.23}, ...]
//
// Returns:
//   - HTTP 400 for invalid JSON
//   - HTTP 500 for storage errors
//   - HTTP 200 on successful batch save
func updatesHandler(storage storage.MetricsStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req []models.Metric
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Log.Error("cannot decode request JSON body", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := storage.SaveAll(r.Context(), req); err != nil {
			logger.Log.Error("cannot save metrics to storage", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

// getAllMetricsHandler returns a handler for retrieving all metrics as HTML.
//
// Returns an HTML page displaying all gauge and counter metrics with their values.
//
// Returns:
//   - HTTP 500 for storage errors
//   - HTTP 200 with HTML content
func getAllMetricsHandler(storage storage.MetricsStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := "<!DOCTYPE html><html><body>"

		gauges, err := storage.GetAllByType(r.Context(), models.GaugeMetricName)
		if err != nil {
			logger.Log.Errorf("Can not get all gauge metrics: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		for _, value := range gauges {
			page += fmt.Sprintf("\n<p>%s - %f</p>", value.ID, *value.Value)
		}

		counters, err := storage.GetAllByType(r.Context(), models.CounterMetricName)
		if err != nil {
			logger.Log.Errorf("Can not get all counter metrics: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		for _, value := range counters {
			page += fmt.Sprintf("\n<p>%s - %d</p>", value.ID, *value.Delta)
		}

		page += "\n</body></html>"

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page))
	}
}

// updateMetricHandler returns a handler for updating metrics via JSON request.
//
// Expected JSON request body format: Metric object
//
//	{"id": "metricName", "mType": "gauge|counter", "value": 1.23, "delta": 42}
//
// Returns:
//   - HTTP 400 for invalid JSON or save errors
//   - HTTP 200 with updated metric data in JSON format
func updateMetricHandler(storage storage.MetricsStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.Metric

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
