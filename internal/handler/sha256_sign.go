package handler

import (
	"bytes"
	"io"
	"net/http"

	"github.com/DenisPavlov/monitoring/internal/util"
)

// SHA256HeaderName is the HTTP header name used for SHA256 signature verification.
// Both requests and responses use this header for signature transmission.
const SHA256HeaderName = "HashSHA256"

// respWriter is a wrapper around http.ResponseWriter that automatically
// signs response bodies with SHA256 hash when a cryptographic key is provided.
//
// It implements the http.ResponseWriter interface and transparently
// adds the SHA256 signature header to responses.
type respWriter struct {
	http.ResponseWriter
	key string
}

// Write writes the data to the underlying ResponseWriter and adds a SHA256
// signature header calculated from the response body and the provided key.
//
// The signature is calculated as SHA256(key + body) and set in the
// SHA256HeaderName header.
//
// Parameters:
//   - b: response body bytes to write and sign
//
// Returns:
//   - int: number of bytes written
//   - error: any write error that occurred
func (w *respWriter) Write(b []byte) (int, error) {
	respSign := util.GetHexSHA256(w.key, b)
	w.Header().Set(SHA256HeaderName, respSign)
	return w.ResponseWriter.Write(b)
}

// SHA256SignMiddleware provides middleware for SHA256 request signature verification
// and response signature generation.
//
// The middleware performs two functions:
//  1. Request Verification: For incoming requests with SHA256HeaderName header,
//     it verifies that the request body matches the provided signature.
//     Returns HTTP 400 if verification fails.
//  2. Response Signing: For outgoing responses, it automatically signs the
//     response body with SHA256 when a key is provided.
//
// Signature format: SHA256(key + body)
//
// Parameters:
//   - key: Cryptographic key used for both verification and signing.
//     If empty, the middleware will skip signature processing.
//
// Returns:
//   - func(http.Handler) http.Handler: Chi middleware function
//
// Usage:
//
//	router := chi.NewRouter()
//	router.Use(SHA256SignMiddleware("your-secret-key"))
//
// The middleware handles both scenarios:
//   - Requests with HashSHA256 header: verifies the signature
//   - Requests without HashSHA256 header: signs the response
//
// Security Note: The key should be kept secret and shared between client and server.
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
