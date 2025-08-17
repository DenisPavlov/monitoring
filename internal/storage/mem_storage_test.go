package storage

import (
	"context"
	"testing"

	"github.com/DenisPavlov/monitoring/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewMemStorage(t *testing.T) {
	s := NewMemStorage()
	assert.NotNil(t, s)
}

func TestMemStorage_Add(t *testing.T) {
	ctx := context.Background()
	s := NewMemStorage()
	value1 := int64(1)
	m := models.Metric{
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
