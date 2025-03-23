package client

import (
	"bytes"
	"encoding/json"
	"github.com/DenisPavlov/monitoring/internal/models"
	"net/http"
)

func postMetric(host string, metrics models.Metrics) error {

	metricBytes, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://"+host+"/update", "application/json", bytes.NewBuffer(metricBytes))
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
