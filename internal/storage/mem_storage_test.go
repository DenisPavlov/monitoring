package storage

import (
	"context"
	"github.com/DenisPavlov/monitoring/internal/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMemStorage(t *testing.T) {
	s := NewMemStorage()
	assert.NotNil(t, s)
}

func TestMemStorage_Add(t *testing.T) {
	ctx := context.Background()
	s := NewMemStorage()
	value1 := int64(1)
	m := models.Metrics{
		ID:    "m1",
		MType: "counter",
		Delta: &value1,
	}
	err := s.Save(ctx, &m)
	assert.NoError(t, err)
	actual, err := s.GetByTypeAndID(ctx, "m1", "counter")
	assert.NoError(t, err)
	assert.Equal(t, m, actual)

	value2 := int64(1)
	m.Delta = &value2
	err = s.Save(ctx, &m)
	assert.NoError(t, err)
	actual, err = s.GetByTypeAndID(ctx, "m1", "counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), *actual.Delta)
}

//func TestMemStorage_All(t *testing.T) {
//	s := NewMemStorage()
//	_ = s.AddCounter("c1", 1)
//	_ = s.AddCounter("c2", 1)
//	_ = s.AddGauge("g1", 1.01)
//	_ = s.AddGauge("g2", 1.01)
//
//	gRes := s.AllGauges()
//	assert.True(t, reflect.DeepEqual(gRes, map[string]float64{"g1": 1.01, "g2": 1.01}))
//
//	cRes := s.AllCounters()
//	assert.True(t, reflect.DeepEqual(cRes, map[string]int64{"c1": 1, "c2": 1}))
//}
