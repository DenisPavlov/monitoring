package main

import (
	"github.com/DenisPavlov/monitoring/internal/client"
	"github.com/DenisPavlov/monitoring/internal/logger"
	"github.com/DenisPavlov/monitoring/internal/service"
	"log"
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
	var (
		counter = make(map[string]int64)
		gauges  map[string]float64
	)
	count := 1

	for {
		if count%flagPollInterval == 0 {
			gauges = metrics.Gauge()
			counter = metrics.Count(counter)
		}

		if count%flagReportInterval == 0 {
			if err := client.PostMetrics(flagRunAddr, counter, gauges); err != nil {
				logger.Log.Error(err)
			}
		}

		count++
		time.Sleep(time.Second)
	}
}
