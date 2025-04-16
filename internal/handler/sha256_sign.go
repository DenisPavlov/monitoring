package handler

import (
	"bytes"
	"github.com/DenisPavlov/monitoring/internal/util"
	"io"
	"net/http"
)

const SHA256HeaderName = "HashSHA256"

type respWriter struct {
	http.ResponseWriter
	key string
}

func (w *respWriter) Write(b []byte) (int, error) {
	respSign := util.GetHexSHA256(w.key, b)
	w.Header().Set(SHA256HeaderName, respSign)
	return w.ResponseWriter.Write(b)
}

func SHA256SignMiddleware(key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hashValue := r.Header.Get(SHA256HeaderName)

			if hashValue != "" {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				sign := util.GetHexSHA256(key, body)

				if sign != hashValue {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				r.Body = io.NopCloser(bytes.NewBuffer(body))
				next.ServeHTTP(w, r)
			} else {
				rw := &respWriter{ResponseWriter: w}
				next.ServeHTTP(rw, r)
			}
		})
	}
}
