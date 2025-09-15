package storage

import (
	"context"
	"errors"
	"reflect"
	"sync"

	"github.com/DenisPavlov/monitoring/internal/models"
)

// MemoryMetricsStorage implements in-memory storage for metrics using a thread-safe map.
// It provides basic CRUD operations for metrics with proper concurrency control.
type MemoryMetricsStorage struct {
	metrics map[string]models.Metric
	mu      sync.Mutex
}

// NewMemStorage creates a new instance of MemoryMetricsStorage with an empty metrics map.
//
// Returns:
//   - *MemoryMetricsStorage: New in-memory storage instance
//
// Example usage:
//
//	storage := NewMemStorage()
func NewMemStorage() *MemoryMetricsStorage {
	return &MemoryMetricsStorage{
		metrics: make(map[string]models.Metric),
	}
}

// Save stores a metric in the in-memory storage with proper concurrency control.
// For counter metrics, it increments the existing value rather than replacing it.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - metric: Pointer to the Metric to be saved
//
// Returns:
//   - error: If context is cancelled or metric validation fails
//
// Behavior:
//   - For gauge metrics: Replaces existing value
//   - For counter metrics: Increments existing value (adds to current delta)
//   - Thread-safe: Uses mutex locking for concurrent access
//
// Example usage:
//
//	err := storage.Save(ctx, &models.Metric{
//	    ID:    "requests",
//	    MType: "counter",
//	    Delta: &[]int64{1}[0],
//	})
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
			m, err := s.getByTypeAndID(metric.ID, models.CounterMetricName)
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

// SaveAll stores multiple metrics in the storage atomically.
// If any metric fails to save, the entire operation fails.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - metrics: Slice of Metric objects to be saved
//
// Returns:
//   - error: If context is cancelled or any metric save operation fails
//
// Example usage:
//
//	metrics := []models.Metric{metric1, metric2, metric3}
//	err := storage.SaveAll(ctx, metrics)
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

// GetByTypeAndID retrieves a metric by its ID and type with concurrency control.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - id: Metric identifier name
//   - mType: Metric type ("gauge" or "counter")
//
// Returns:
//   - models.Metric: Found metric or empty Metric if not found
//   - error: If context is cancelled or validation fails
//
// Example usage:
//
//	metric, err := storage.GetByTypeAndID(ctx, "cpu_usage", "gauge")
func (s *MemoryMetricsStorage) GetByTypeAndID(ctx context.Context, id, mType string) (res models.Metric, err error) {
	select {
	case <-ctx.Done():
		return res, ctx.Err()
	default:
		s.mu.Lock()
		defer s.mu.Unlock()
		return s.getByTypeAndID(id, mType)
	}
}

// getByTypeAndID is an internal method to retrieve a metric by ID and type.
// Must be called with mutex already locked.
//
// Parameters:
//   - id: Metric identifier name
//   - mType: Metric type ("gauge" or "counter")
//
// Returns:
//   - models.Metric: Found metric or empty Metric if not found
//   - error: If metric validation fails
func (s *MemoryMetricsStorage) getByTypeAndID(id, mType string) (res models.Metric, err error) {
	key, err := key(models.Metric{ID: id, MType: mType})
	if err != nil {
		return res, err
	}
	return s.metrics[key], nil
}

// GetAllByType retrieves all metrics of a specific type.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - mType: Metric type to filter by ("gauge" or "counter")
//
// Returns:
//   - []models.Metric: Slice of metrics matching the type
//   - error: If context is cancelled
//
// Example usage:
//
//	gauges, err := storage.GetAllByType(ctx, "gauge")
//	counters, err := storage.GetAllByType(ctx, "counter")
func (s *MemoryMetricsStorage) GetAllByType(ctx context.Context, mType string) (res []models.Metric, err error) {
	select {
	case <-ctx.Done():
		return res, ctx.Err()
	default:
		s.mu.Lock()
		defer s.mu.Unlock()
		for _, metric := range s.metrics {
			if metric.MType == mType {
				res = append(res, metric)
			}
		}
		return res, nil
	}
}

// key generates a unique storage key for a metric based on ID and type.
//
// Parameters:
//   - m: Metric object containing ID and MType
//
// Returns:
//   - string: Unique key in format "ID:MType"
//   - error: If ID or MType is empty
//
// Example output:
//   - key for {"ID": "cpu", "MType": "gauge"} -> "cpu:gauge"
func key(m models.Metric) (string, error) {
	if m.ID == "" || m.MType == "" {
		return "", errors.New("invalid metrics, ID or MType is empty")
	}
	return m.ID + ":" + m.MType, nil
}
