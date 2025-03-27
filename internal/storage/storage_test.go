package storage

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNewMemStorage(t *testing.T) {
	s := NewMemStorage(false, "")
	assert.NotNil(t, s)
}

func TestMemStorage_Add(t *testing.T) {
	s := NewMemStorage(false, "")
	_ = s.AddCounter("m1", 1)
	_ = s.AddGauge("m1", 1.01)

	cVal, ok := s.Counter("m1")
	assert.True(t, ok)
	assert.Equal(t, cVal, int64(1))

	gVal, ok := s.Gauge("m1")
	assert.True(t, ok)
	assert.Equal(t, gVal, 1.01)
}

func TestMemStorage_All(t *testing.T) {
	s := NewMemStorage(false, "")
	_ = s.AddCounter("c1", 1)
	_ = s.AddCounter("c2", 1)
	_ = s.AddGauge("g1", 1.01)
	_ = s.AddGauge("g2", 1.01)

	gRes := s.AllGauges()
	assert.True(t, reflect.DeepEqual(gRes, map[string]float64{"g1": 1.01, "g2": 1.01}))

	cRes := s.AllCounters()
	assert.True(t, reflect.DeepEqual(cRes, map[string]int64{"c1": 1, "c2": 1}))
}
