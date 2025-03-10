package client

import (
	"fmt"
	"net/http"
)

func postMetric(url string) error {
	resp, err := http.Post(url, "text/plain", http.NoBody)
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
		url := fmt.Sprintf("http://"+host+"/update/gauge/%s/%f", name, value)
		err := postMetric(url)
		if err != nil {
			return err
		}
	}

	for name, value := range counts {
		url := fmt.Sprintf(host+"/update/counter/%s/%d", name, value)
		err := postMetric(url)
		if err != nil {
			return err
		}
	}
	return nil
}
