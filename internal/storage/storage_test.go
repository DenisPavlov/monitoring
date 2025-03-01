package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMemStorage(t *testing.T) {
	s := NewMemStorage()
	assert.NotNil(t, s)
}

func TestMemStorage_Add(t *testing.T) {
	s := NewMemStorage()
	s.AddCounter("m1", 1)
	s.AddGauge("m1", 1.01)

	assert.Equal(t, s.Counter("m1"), int64(1))
	assert.Equal(t, s.Gauge("m1"), 1.01)
}
