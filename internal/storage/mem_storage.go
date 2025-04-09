package storage

import (
	"context"
	"errors"
	"github.com/DenisPavlov/monitoring/internal/models"
	"reflect"
)

type MemStorage struct {
	metrics map[string]models.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]models.Metrics),
	}
}

func (s MemStorage) Save(ctx context.Context, metric *models.Metrics) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		key, err := key(*metric)
		if err != nil {
			return err
		}
		switch metric.MType {
		case models.GaugeMetricName:
			s.metrics[key] = *metric
		case models.CounterMetricName:
			m, err := s.GetByTypeAndID(ctx, metric.ID, models.CounterMetricName)
			if err != nil {
				return err
			}
			if !reflect.DeepEqual(m, models.Metrics{}) {
				*metric.Delta = *metric.Delta + *m.Delta
			}
			s.metrics[key] = *metric
		}
		return nil
	}
}

/*
//		if metrics.Delta == nil {
//			return ErrWrongValue
//		}
//		if err := storage.AddCounter(metrics.ID, *metrics.Delta); err != nil {
//			return err
//		}
//		updatedValue, _ := storage.Counter(metrics.ID)
//		*metrics.Delta = updatedValue
*/

func (s MemStorage) SaveAll(ctx context.Context, metrics []models.Metrics) error {
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

func (s MemStorage) GetByTypeAndID(ctx context.Context, ID string, mType string) (res models.Metrics, err error) {
	select {
	case <-ctx.Done():
		return res, ctx.Err()
	default:
		key, err := key(models.Metrics{ID: ID, MType: mType})
		if err != nil {
			return res, err
		}
		return s.metrics[key], nil
	}
}
func (s MemStorage) GetAllByType(ctx context.Context, mType string) (res []models.Metrics, err error) {
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

func key(m models.Metrics) (string, error) {
	if m.ID == "" || m.MType == "" {
		return "", errors.New("invalid metrics, ID or MType is empty")
	}
	return m.ID + ":" + m.MType, nil
}
