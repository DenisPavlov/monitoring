package logger

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

var Log = log.New()

func Initialize(level string, env string) error {
	lvl, err := log.ParseLevel(level)
	if err != nil {
		return err
	}
	Log.SetLevel(lvl)
	Log.SetOutput(os.Stdout)

	switch env {
	case "production":
		Log.SetFormatter(&log.JSONFormatter{})
	case "development":
		Log.SetFormatter(&log.TextFormatter{})
	}
	return nil
}

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rd := &responseData{status: http.StatusOK}
		lrw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   rd,
		}
		next.ServeHTTP(&lrw, r)

		Log.WithFields(log.Fields{
			"method":   r.Method,
			"path":     r.URL.Path,
			"duration": time.Since(start),
			"status":   rd.status,
			"size":     rd.size,
		}).Info("got incoming HTTP request")
	})
}
