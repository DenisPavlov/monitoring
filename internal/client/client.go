package client

import (
	"fmt"
	"net/http"
)

func postMetric(client *http.Client, url string) error {
	_, err := client.Post(url, "text/plain", http.NoBody)
	if err != nil {
		return err
	}
	return nil
}

func PostMetrics(client *http.Client, counts map[string]int, gauges map[string]float64) error {
	for name, value := range gauges {
		url := fmt.Sprintf("http://localhost:8080/update/gauge/%s/%f", name, value)
		err := postMetric(client, url)
		if err != nil {
			return err
		}
	}

	for name, value := range counts {
		url := fmt.Sprintf("http://localhost:8080/update/counter/%s/%d", name, value)
		err := postMetric(client, url)
		if err != nil {
			return err
		}
	}
	return nil
}
