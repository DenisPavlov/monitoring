package handler

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/DenisPavlov/monitoring/internal/logger"
)

var _ http.ResponseWriter = (*compressWriter)(nil)

// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: nil,
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write writes the data to the connection as part of an HTTP reply.
// For supported content types (application/json, text/html), it compresses
// the data using gzip. For other content types, it writes uncompressed data.
func (c *compressWriter) Write(p []byte) (int, error) {
	// поверить, что тип контента application/json или text/html
	contentType := c.w.Header().Get("Content-Type")
	supportsContentType :=
		strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html")

	if supportsContentType {
		c.zw = gzip.NewWriter(c.w)
		c.Header().Set("Content-Encoding", "gzip")
		return c.zw.Write(p)
	} else {
		return c.w.Write(p)
	}
}

func (c *compressWriter) WriteHeader(statusCode int) {
	c.w.WriteHeader(statusCode)
}

// Close closes the gzip.Writer and flushes any buffered data.
// This method should be called to ensure all compressed data is sent to the client.
func (c *compressWriter) Close() error {
	if c.zw != nil {
		return c.zw.Close()
	}
	return nil
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read reads decompressed data from the underlying gzip stream.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close closes both the original reader and the gzip reader.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// GzipMiddleware provides HTTP middleware for gzip compression and decompression.
//
// The middleware performs:
//   - Response compression: Compresses responses with gzip for clients that
//     support it (Accept-Encoding: gzip) and for supported content types
//   - Request decompression: Decompresses gzip-encoded request bodies from clients
//     (Content-Encoding: gzip)
//
// Supported content types for compression:
//   - application/json
//   - text/html
//
// Usage:
//
//	router := mux.NewRouter()
//	router.Use(GzipMiddleware)
//
// The middleware automatically handles the compression and decompression
// transparently for the wrapped handlers.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		ow := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := newCompressWriter(w)
			// меняем оригинальный http.ResponseWriter на новый
			ow = cw
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer func() {
				if err := cw.Close(); err != nil {
					logger.Log.Errorf("failed to close compress writer: %v", err)
				}
			}()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer func() {
				if err := cr.Close(); err != nil {
					logger.Log.Errorf("failed to close compressreader: %v", err)
				}
			}()
		}

		// передаём управление хендлеру
		next.ServeHTTP(ow, r)
	})
}
