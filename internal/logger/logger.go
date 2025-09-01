// Package logger provides logging functionality for the application
// including HTTP request logging middleware and structured logging setup.
package logger

import (
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

var _ http.ResponseWriter = (*loggingResponseWriter)(nil)

// Log is the global logger instance used throughout the application.
// It is initialized with default settings and can be configured using Initialize().
var Log = log.New()

// Initialize configures the global logger with the specified log level and environment.
//
// Parameters:
//   - level: Log level string ("debug", "info", "warn", "error", "fatal", "panic")
//   - env: Environment string ("production" or "development")
//
// Returns:
//   - error: If the log level cannot be parsed
//
// Configuration:
//   - Production environment: Uses JSON formatter for structured logging
//   - Development environment: Uses text formatter for human-readable output
//   - Log output is always set to stdout
//
// Example usage:
//
//	err := logger.Initialize("info", "production")
//	if err != nil {
//	    log.Fatal("Failed to initialize logger:", err)
//	}
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

// responseData holds metadata about the HTTP response for logging purposes.
type (
	responseData struct {
		status int
		size   int
	}

	// loggingResponseWriter wraps http.ResponseWriter to capture response details.
	// It tracks HTTP status code and response size for comprehensive request logging.
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Write delegates to the underlying ResponseWriter while tracking the response size.
//
// Parameters:
//   - b: Bytes to write to the response
//
// Returns:
//   - int: Number of bytes written
//   - error: Any error that occurred during writing
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader delegates to the underlying ResponseWriter while capturing the status code.
//
// Parameters:
//   - statusCode: HTTP status code to set
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// RequestLogger provides HTTP middleware for comprehensive request logging.
//
// The middleware logs:
//   - Incoming requests: method, path, and content length
//   - Completed requests: method, path, duration, status code, and response size
//
// Returns:
//   - func(http.Handler) http.Handler: Chi middleware function
//
// Usage:
//
//	router := chi.NewRouter()
//	router.Use(logger.RequestLogger)
//
// Example log output:
//   - Incoming: "method=GET path=/api/metrics contentLength=0"
//   - Completed: "method=GET path=/api/metrics duration=12.5ms status=200 size=1024"
//
// The middleware uses the global Log instance and respects its configuration.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		Log.WithFields(log.Fields{
			"method":        r.Method,
			"path":          r.URL.Path,
			"contentLength": r.ContentLength,
		}).Info("got incoming HTTP request")

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
		}).Info("completed HTTP request")
	})
}
