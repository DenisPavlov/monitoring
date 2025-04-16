package metrics

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCount(t *testing.T) {
	m := make(map[string]int64)
	Count(m)
	assert.Equal(t, m["PollCount"], int64(1))
}

func TestGauge(t *testing.T) {
	m := Gauge()
	assert.NotZero(t, m["RandomValue"])
}
