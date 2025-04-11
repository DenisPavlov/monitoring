package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/DenisPavlov/monitoring/internal/logger"
	"github.com/DenisPavlov/monitoring/internal/models"
	"math"
	"net/http"
	"time"
)

func postMetric(host string, metrics []models.Metrics) error {

	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	err := json.NewEncoder(gzipWriter).Encode(metrics)
	if err != nil {
		return err
	}
	err = gzipWriter.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "http://"+host+"/updates/", &buffer)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	retries := 4
	var resp *http.Response

	for i := 0; i < retries; i++ {
		logger.Log.Infof("Posting metrics to %s with attention %d", host, i)
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		if shouldRetry(resp) {
			time.Sleep(backoff(i))
		} else {
			if err := resp.Body.Close(); err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

func PostMetrics(host string, counts map[string]int64, gauges map[string]float64) error {

	var metrics []models.Metrics
	for name, value := range gauges {
		metric := models.Metrics{
			ID:    name,
			MType: "gauge",
			Value: &value,
		}
		metrics = append(metrics, metric)
	}

	for name, value := range counts {
		metric := models.Metrics{
			ID:    name,
			MType: "counter",
			Delta: &value,
		}
		metrics = append(metrics, metric)
	}
	err := postMetric(host, metrics)
	if err != nil {
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

func backoff(retries int) time.Duration {
	return time.Duration(math.Pow(2, float64(retries))) * time.Second
}
