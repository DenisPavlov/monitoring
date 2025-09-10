package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/DenisPavlov/monitoring/cmd/agent/config"
	"github.com/DenisPavlov/monitoring/internal/build/info"
	"github.com/DenisPavlov/monitoring/internal/client"
	"github.com/DenisPavlov/monitoring/internal/models"
	"github.com/DenisPavlov/monitoring/internal/service"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	info.PrintBuildInfo(buildVersion, buildDate, buildCommit)

	if err := config.ParseFlags(); err != nil {
		log.Fatal(err)
	}
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var wg sync.WaitGroup

	metricsChan := make(chan []models.Metric)
	wg.Add(1)
	go func() {
		defer wg.Done()
		collectAndSend(ctx, time.Duration(config.FlagPollInterval)*time.Second, metricsChan)
	}()

	reportChan := make(chan []models.Metric)
	wg.Add(1)
	go func() {
		defer wg.Done()
		collectReport(ctx, time.Duration(config.FlagReportInterval)*time.Second, metricsChan, reportChan)
	}()

	wg.Add(config.FlagRateLimit)
	for i := 0; i < config.FlagRateLimit; i++ {
		go func(workerID int) {
			defer wg.Done()
			postMetricsWorker(ctx, workerID, reportChan)
		}(i)
	}

	wg.Wait()

	return nil
}

func collectAndSend(ctx context.Context, interval time.Duration, out chan<- []models.Metric) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	counters := make(map[string]int64)

	for {
		select {
		case <-ctx.Done():
			close(out)
			return
		case <-ticker.C:
			log.Printf("Collecting metrics")
			var metricsBatch []models.Metric

			for name, value := range metrics.Gauge() {
				metricsBatch = append(metricsBatch, models.Metric{
					ID:    name,
					MType: "gauge",
					Value: &value,
				})
			}

			for name, value := range metrics.AdditionalGauge() {
				metricsBatch = append(metricsBatch, models.Metric{
					ID:    name,
					MType: "gauge",
					Value: &value,
				})
			}

			counters = metrics.Count(counters)
			for name, value := range counters {
				metricsBatch = append(metricsBatch, models.Metric{
					ID:    name,
					MType: "counter",
					Delta: &value,
				})
			}

			out <- metricsBatch
		}
	}
}

func postMetricsWorker(ctx context.Context, workerID int, in <-chan []models.Metric) {
	for {
		select {
		case <-ctx.Done():
			return
		case metric, ok := <-in:
			if !ok {
				return
			}
			log.Printf("[Worker %d] Sending %d metrics", workerID, len(metric))
			if err := client.PostMetricsBatch(ctx, config.FlagRunAddr, config.FlagKey, metric); err != nil {
				log.Printf("[Worker %d] Error sending metrics: %v", workerID, err)
			}
		}
	}
}

func collectReport(ctx context.Context, interval time.Duration, in <-chan []models.Metric, out chan<- []models.Metric) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var currentMetrics []models.Metric
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Printf("Sending metrics: %v", currentMetrics)
			out <- currentMetrics
		case metric, ok := <-in:
			if !ok {
				return
			}
			currentMetrics = metric
		}
	}
}
