package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/DenisPavlov/monitoring/internal/handler"
	"github.com/DenisPavlov/monitoring/internal/logger"
	"github.com/DenisPavlov/monitoring/internal/models"
	"github.com/DenisPavlov/monitoring/internal/util"
	"net/http"
	"time"
)

func postMetric(host, signKey string, metrics []models.Metric) error {

	strReqBody, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	if _, err := gzipWriter.Write(strReqBody); err != nil {
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

func PostMetrics(host, signKey string, counters map[string]int64, gauges map[string]float64) error {

	var metrics []models.Metric
	for name, value := range gauges {
		metric := models.Metric{
			ID:    name,
			MType: "gauge",
			Value: &value,
		}
		metrics = append(metrics, metric)
	}

	for name, value := range counters {
		metric := models.Metric{
			ID:    name,
			MType: "counter",
			Delta: &value,
		}
		metrics = append(metrics, metric)
	}
	err := postMetric(host, signKey, metrics)
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

func signRequest(req *http.Request, body []byte, key string) {
	sign := util.GetHexSHA256(key, body)
	req.Header.Set(handler.SHA256HeaderName, sign)
}
