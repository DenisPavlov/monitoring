package util

import (
	"math"
	"time"
)

func Backoff(retries int) time.Duration {
	return time.Duration(math.Pow(2, float64(retries))) * time.Second
}
