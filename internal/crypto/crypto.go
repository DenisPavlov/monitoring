// Package crypto provides utilities for asymmetric encryption using RSA keys.
// It supports loading public/private keys from PEM files, encrypting data with
// public keys, and decrypting data with private keys.
package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// LoadPublicKey loads an RSA public key from a PEM-encoded file.
// The file should contain a PKCS#1 public key in PEM format with "RSA PUBLIC KEY" header.
//
// Example PEM format:
// -----BEGIN RSA PUBLIC KEY-----
// ... base64 encoded key ...
// -----END RSA PUBLIC KEY-----
//
// Parameters:
//   - filename: Path to the public key file
//
// Returns:
//   - *rsa.PublicKey: The loaded public key
//   - error: Any error encountered during loading or parsing
//
// Errors:
//   - If the file cannot be read
//   - If the PEM block cannot be decoded
//   - If the key format is invalid or not RSA
func LoadPublicKey(filename string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing public key")
	}

	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return publicKey, nil
}

// LoadPrivateKey loads an RSA private key from a PEM-encoded file.
// Supports both PKCS#1 and PKCS#8 formats. The function will automatically
// detect the format and parse accordingly.
//
// Supported PEM headers:
//   - PKCS#1: "RSA PRIVATE KEY"
//   - PKCS#8: "PRIVATE KEY"
//
// Parameters:
//   - filename: Path to the private key file
//
// Returns:
//   - *rsa.PrivateKey: The loaded private key
//   - error: Any error encountered during loading or parsing
//
// Errors:
//   - If the file cannot be read
//   - If the PEM block cannot be decoded
//   - If the key format is invalid or not RSA
//   - If the key cannot be parsed in either PKCS#1 or PKCS#8 format
func LoadPrivateKey(filename string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Попробуем PKCS8 format
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		privateKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not an RSA private key")
		}
		return privateKey, nil
	}

	return privateKey, nil
}

// EncryptData encrypts data using an RSA public key with PKCS#1 v1.5 padding.
// For data larger than the RSA key size, it automatically splits the data into
// chunks and encrypts each chunk separately, then concatenates the results.
//
// The maximum chunk size is calculated as: key_size - 42 bytes (for OAEP padding overhead).
// This ensures compatibility with standard RSA encryption limitations.
//
// Parameters:
//   - publicKey: The RSA public key to use for encryption
//   - data: The plaintext data to encrypt
//
// Returns:
//   - []byte: The encrypted data
//   - error: Any error encountered during encryption
//
// Errors:
//   - If encryption fails for any chunk
//   - If the public key is invalid
func EncryptData(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	// RSA может шифровать только ограниченное количество данных
	// Разделяем данные на chunks если нужно
	maxChunkSize := publicKey.Size() - 42 // Для OAEP padding
	if len(data) <= maxChunkSize {
		return rsa.EncryptPKCS1v15(rand.Reader, publicKey, data)
	}

	var encrypted []byte
	for i := 0; i < len(data); i += maxChunkSize {
		end := i + maxChunkSize
		if end > len(data) {
			end = len(data)
		}
		chunk := data[i:end]

		encryptedChunk, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, chunk)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt chunk: %w", err)
		}
		encrypted = append(encrypted, encryptedChunk...)
	}

	return encrypted, nil
}

// DecryptData decrypts data using an RSA private key with PKCS#1 v1.5 padding.
// Expects the encrypted data to be in chunks that are multiples of the RSA key size.
// Decrypts each chunk separately and concatenates the results.
//
// Parameters:
//   - privateKey: The RSA private key to use for decryption
//   - encryptedData: The encrypted data to decrypt
//
// Returns:
//   - []byte: The decrypted plaintext data
//   - error: Any error encountered during decryption
//
// Errors:
//   - If the encrypted data length is not a multiple of the key size
//   - If decryption fails for any chunk
//   - If the private key is invalid
//
// Note: The encrypted data should have been produced by EncryptData function
// to ensure proper chunk sizing and formatting.
func DecryptData(privateKey *rsa.PrivateKey, encryptedData []byte) ([]byte, error) {
	keySize := privateKey.Size()
	if len(encryptedData)%keySize != 0 {
		return nil, errors.New("encrypted data length is not multiple of key size")
	}

	var decrypted []byte
	for i := 0; i < len(encryptedData); i += keySize {
		end := i + keySize
		if end > len(encryptedData) {
			end = len(encryptedData)
		}
		chunk := encryptedData[i:end]

		decryptedChunk, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, chunk)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt chunk: %w", err)
		}
		decrypted = append(decrypted, decryptedChunk...)
	}

	return decrypted, nil
}
