package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/DenisPavlov/monitoring/internal/models"
	"net/http"
)

func postMetric(host string, metrics models.Metrics) error {

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

	req, err := http.NewRequest("POST", "http://"+host+"/update", &buffer)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}
	if err := resp.Body.Close(); err != nil {
		return err
	}
	return nil
}

func PostMetrics(host string, counts map[string]int64, gauges map[string]float64) error {
	for name, value := range gauges {
		metrics := models.Metrics{
			ID:    name,
			MType: "gauge",
			Value: &value,
		}

		err := postMetric(host, metrics)
		if err != nil {
			return err
		}
	}

	for name, value := range counts {
		metrics := models.Metrics{
			ID:    name,
			MType: "counter",
			Delta: &value,
		}
		err := postMetric(host, metrics)
		if err != nil {
			return err
		}
	}
	return nil
}
