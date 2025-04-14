package storage

import (
	"context"
	"errors"
	"github.com/DenisPavlov/monitoring/internal/models"
	"reflect"
	"sync"
)

type MemoryMetricsStorage struct {
	mu      sync.Mutex
	metrics map[string]models.Metric
}

func NewMemStorage() *MemoryMetricsStorage {
	return &MemoryMetricsStorage{
		metrics: make(map[string]models.Metric),
	}
}

func (s *MemoryMetricsStorage) Save(ctx context.Context, metric *models.Metric) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		key, err := key(*metric)
		if err != nil {
			return err
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		switch metric.MType {
		case models.GaugeMetricName:
			s.metrics[key] = *metric
		case models.CounterMetricName:
			m, err := s.GetByTypeAndID(ctx, metric.ID, models.CounterMetricName)
			if err != nil {
				return err
			}
			if !reflect.DeepEqual(m, models.Metric{}) {
				*metric.Delta = *metric.Delta + *m.Delta
			}
			s.metrics[key] = *metric
		}
		return nil
	}
}

func (s *MemoryMetricsStorage) SaveAll(ctx context.Context, metrics []models.Metric) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		for _, metric := range metrics {
			err := s.Save(ctx, &metric)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func (s *MemoryMetricsStorage) GetByTypeAndID(ctx context.Context, ID, mType string) (res models.Metric, err error) {
	select {
	case <-ctx.Done():
		return res, ctx.Err()
	default:
		key, err := key(models.Metric{ID: ID, MType: mType})
		if err != nil {
			return res, err
		}
		return s.metrics[key], nil
	}
}
func (s *MemoryMetricsStorage) GetAllByType(ctx context.Context, mType string) (res []models.Metric, err error) {
	select {
	case <-ctx.Done():
		return res, ctx.Err()
	default:
		for _, metric := range s.metrics {
			if metric.MType == mType {
				res = append(res, metric)
			}
		}
		return res, nil
	}
}

func key(m models.Metric) (string, error) {
	if m.ID == "" || m.MType == "" {
		return "", errors.New("invalid metrics, ID or MType is empty")
	}
	return m.ID + ":" + m.MType, nil
}
