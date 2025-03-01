package main

import (
	"github.com/DenisPavlov/monitoring/internal/client"
	"github.com/DenisPavlov/monitoring/internal/measure"
	"log"
	"time"
)

const (
	pollIntervalSec   = 2
	reportIntervalSec = 10
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	counts := make(map[string]int64)
	var gauges map[string]float64
	count := 1

	for {
		if count%pollIntervalSec == 0 {
			gauges = measure.Gauge()
			counts = measure.Count(counts)
		}

		if count%reportIntervalSec == 0 {
			if err := client.PostMetrics(counts, gauges); err != nil {
				return err
			}
		}

		count++
		time.Sleep(time.Second)
	}
}
