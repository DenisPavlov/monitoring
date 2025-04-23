package main

import (
	"github.com/DenisPavlov/monitoring/internal/client"
	"github.com/DenisPavlov/monitoring/internal/models"
	"github.com/DenisPavlov/monitoring/internal/service"
	"log"
	"sync"
	"time"
)

func main() {
	if err := parseFlags(); err != nil {
		log.Fatal(err)
	}
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var wg sync.WaitGroup
	wg.Add(1)

	commonChan := make(chan []models.Metric)
	additionalChan := make(chan []models.Metric)
	go poolCommon(flagPollInterval, commonChan)
	go poolAdditional(flagPollInterval, additionalChan)

	reportChan := make(chan []models.Metric)
	go report(flagReportInterval, reportChan, commonChan, additionalChan)

	client.PostMetricsAsync(flagRunAddr, flagKey, flagRateLimit, reportChan)

	wg.Wait()

	return nil
}

func report(reportInterval int, reportChan chan<- []models.Metric, metricsChan ...chan []models.Metric) {
	for _, ch := range metricsChan {
		go func() {
			ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
			for m := range ch {
				currentMetrics := m
				select {
				case <-ticker.C:
					log.Printf("Sending metrics: %v", currentMetrics)
					reportChan <- currentMetrics
				default:
					log.Printf("Skip sending metrics: %v", currentMetrics)
				}
			}
		}()
	}
}

func poolCommon(poolInterval int, metricsCh chan<- []models.Metric) {
	ticker := time.NewTicker(time.Duration(poolInterval) * time.Second)
	var (
		gauges   map[string]float64
		counters = make(map[string]int64)
	)

	for range ticker.C {
		log.Println("poolCommon metrics")
		gauges = metrics.Gauge()
		counters = metrics.Count(counters)

		var res []models.Metric
		for name, value := range gauges {
			metric := models.Metric{
				ID:    name,
				MType: "gauge",
				Value: &value,
			}
			res = append(res, metric)
		}

		for name, value := range counters {
			metric := models.Metric{
				ID:    name,
				MType: "counter",
				Delta: &value,
			}
			res = append(res, metric)
		}
		metricsCh <- res
	}
}

func poolAdditional(poolInterval int, metricsCh chan<- []models.Metric) {
	ticker := time.NewTicker(time.Duration(poolInterval) * time.Second)
	for range ticker.C {
		log.Println("poolAdditional metrics")
		gauges := metrics.AdditionalGauge()
		var res []models.Metric
		for name, value := range gauges {
			metric := models.Metric{
				ID:    name,
				MType: "gauge",
				Value: &value,
			}
			res = append(res, metric)
		}
		metricsCh <- res
	}
}
