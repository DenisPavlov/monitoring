package main

import (
	client2 "github.com/DenisPavlov/monitoring/internal/client"
	"github.com/DenisPavlov/monitoring/internal/measure"
	"log"
	"net/http"
	"time"
)

const (
	pollIntervalSec   = 2
	reportIntervalSec = 10
)

func main() {
	if err := run(&http.Client{}); err != nil {
		log.Fatal(err)
	}
}

func run(client *http.Client) error {
	counts := make(map[string]int)
	var gauges map[string]float64
	count := 0

	for {
		if count%pollIntervalSec == 0 {
			gauges = measure.Gauge()
			counts = measure.Count(counts)
		}

		if count%reportIntervalSec == 0 {
			if err := client2.PostMetrics(client, counts, gauges); err != nil {
				return err
			}

		}

		count++
		time.Sleep(time.Second)
	}
}
