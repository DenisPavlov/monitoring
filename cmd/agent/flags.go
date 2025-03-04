package main

import (
	"flag"
)

var flagRunAddr string
var flagReportInterval int
var flagPollInterval int

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "server address and port")
	flag.IntVar(&flagReportInterval, "r", 10, "frequency of sending metrics to the server in seconds")
	flag.IntVar(&flagPollInterval, "p", 2, "frequency of getting runtime metrics in seconds")
	flag.Parse()
}
