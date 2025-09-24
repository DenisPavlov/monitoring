// Package client provides functionality for sending metrics to a monitoring server.
// It includes features like request compression, signing, and retry mechanisms.
package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"time"

	"github.com/DenisPavlov/monitoring/cmd/agent/config"
	"github.com/DenisPavlov/monitoring/internal/crypto"
	"github.com/DenisPavlov/monitoring/internal/handler"
	"github.com/DenisPavlov/monitoring/internal/logger"
	"github.com/DenisPavlov/monitoring/internal/models"
	"github.com/DenisPavlov/monitoring/internal/util"
)

type Client struct {
	publicKey *rsa.PublicKey
}

func NewClient() (*Client, error) {
	var publicKey *rsa.PublicKey

	if config.CryptoKey != "" {
		key, err := crypto.LoadPublicKey(config.CryptoKey)
		if err != nil {
			logger.Log.Errorf("Error loading public key: %v\n", err)
			return nil, err
		}
		publicKey = key
	}
	return &Client{publicKey: publicKey}, nil
}

func (c *Client) postMetric(ctx context.Context, host, signKey string, metrics []models.Metric) error {
	strReqBody, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	if c.publicKey != nil {
		encryptedData, err := crypto.EncryptData(c.publicKey, strReqBody)
		println(string(encryptedData))
		if err == nil {
			strReqBody = encryptedData
		} else {
			logger.Log.Errorf("Error encrypting data: %v\n", err)
			return err
		}
	}

	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	if _, err = gzipWriter.Write(strReqBody); err != nil {
		return err
	}

	err = gzipWriter.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+host+"/updates/", &buffer)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	if signKey != "" {
		signRequest(req, strReqBody, signKey)
	}

	retries := 4
	var resp *http.Response

	for i := 0; i < retries; i++ {
		logger.Log.Infof("Posting metrics to %s with attention %d", host, i)
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		if shouldRetry(resp) {
			time.Sleep(util.Backoff(i))
		} else {
			if err := resp.Body.Close(); err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

// PostMetricsBatch sends a batch of metrics to the monitoring server.
//
// This is the public interface for sending metrics. It wraps the internal
// postMetric function with error logging.
//
// Parameters:
//   - ctx: context for request cancellation and timeout
//   - host: target server address (format: "host:port")
//   - signKey: cryptographic key for request signing
//   - metrics: slice of Metric objects to send
//
// Returns:
//   - error: if the metrics posting operation fails
//
// Example usage:
//
//	metrics := []models.Metric{...}
//	err := client.PostMetricsBatch(ctx, "localhost:8080", "secret-key", metrics)
//	if err != nil {
//	    // handle error
//	}
func (c *Client) PostMetricsBatch(ctx context.Context, host, signKey string, metrics []models.Metric) error {
	if err := c.postMetric(ctx, host, signKey, metrics); err != nil {
		logger.Log.Errorf("Posting metrics failed: %s", err.Error())
		return err
	}
	return nil
}

func shouldRetry(resp *http.Response) bool {
	if resp.StatusCode == http.StatusBadGateway ||
		resp.StatusCode == http.StatusServiceUnavailable ||
		resp.StatusCode == http.StatusGatewayTimeout {
		return true
	}
	return false
}

func signRequest(req *http.Request, body []byte, key string) {
	sign := util.GetHexSHA256(key, body)
	req.Header.Set(handler.SHA256HeaderName, sign)
}
