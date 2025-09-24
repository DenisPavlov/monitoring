package handler

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/DenisPavlov/monitoring/internal/crypto"
)

// PrivateKeyDecodingMiddleware creates an HTTP middleware that decrypts incoming request bodies
// using an RSA private key. This middleware is designed to work with encrypted data sent by
// clients that used the corresponding public key for encryption.
//
// The middleware:
//  1. Reads the entire request body
//  2. Attempts to decrypt it using the provided private key
//  3. Replaces the request body with the decrypted content
//  4. Passes the request to the next handler in the chain
//
// If any error occurs during reading or decryption, the middleware responds with
// HTTP 400 Bad Request and stops further processing.
//
// Parameters:
//   - privateKey: The RSA private key used for decryption. Must match the public key
//     used by the client for encryption.
//
// Returns:
//   - func(next http.Handler) http.Handler: A middleware function that wraps the next handler.
//
// Usage example:
//
//	privateKey, _ := crypto.LoadPrivateKey("private.pem")
//	router := chi.NewRouter()
//	router.Use(handler.PrivateKeyDecodingMiddleware(privateKey))
//	router.Post("/metrics", handler.HandleMetric)
//
// Note:
//   - This middleware should be used when clients encrypt their request bodies using
//     the corresponding public key.
//   - The middleware reads the entire body into memory, so it may not be suitable for
//     very large payloads without additional considerations.
//   - Ensure the private key is properly secured and not exposed in logs or error messages.
//
// Security considerations:
//   - The private key should be stored securely and not hardcoded in the application.
//   - Consider using environment variables or secure key management systems.
//   - Use appropriate key sizes (2048 bits or higher) for RSA keys.
func PrivateKeyDecodingMiddleware(privateKey *rsa.PrivateKey) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			decryptedBody, err := crypto.DecryptData(privateKey, body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(decryptedBody))
			next.ServeHTTP(w, r)
		})
	}
}
