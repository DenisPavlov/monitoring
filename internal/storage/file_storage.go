package storage

import (
	"context"
	"encoding/json"
	"os"

	"github.com/DenisPavlov/monitoring/internal/logger"
	"github.com/DenisPavlov/monitoring/internal/models"
)

// FileMetricsStorage implements MetricsStorage interface with file-based persistence.
// It wraps MemoryMetricsStorage and adds functionality to save metrics to a JSON file.
//
// The storage supports both synchronous and asynchronous file saving modes.
type FileMetricsStorage struct {
	*MemoryMetricsStorage
	needToSaveSync bool
	filename       string
}

// jsonMetrics is an internal struct used for JSON serialization/deserialization
// of metrics data to/from the file system.
type jsonMetrics struct {
	Metrics map[string]models.Metric
}

// NewFileStorage creates a new FileMetricsStorage instance with empty metrics.
//
// Parameters:
//   - needToSaveSync: If true, metrics are saved to file synchronously on each Save operation.
//     If false, metrics are only saved in memory until explicitly saved to file.
//   - filename: Path to the JSON file where metrics will be persisted.
//
// Returns:
//   - *FileMetricsStorage: New file-based storage instance
//
// Example usage:
//
//	storage := NewFileStorage(true, "/tmp/metrics.json")
func NewFileStorage(needToSaveSync bool, filename string) *FileMetricsStorage {
	return &FileMetricsStorage{
		MemoryMetricsStorage: NewMemStorage(),
		needToSaveSync:       needToSaveSync,
		filename:             filename,
	}
}

// InitFromFile creates a new FileMetricsStorage instance and initializes it
// with metrics data loaded from an existing JSON file.
//
// Parameters:
//   - needToSaveSync: If true, enables synchronous file saving on each update
//   - filename: Path to the JSON file to load metrics from
//
// Returns:
//   - *FileMetricsStorage: Storage instance initialized with file data
//   - error: If file reading or JSON unmarshaling fails
//
// Example usage:
//
//	storage, err := InitFromFile(true, "/tmp/metrics.json")
//	if err != nil {
//	    log.Fatal("Failed to initialize from file:", err)
//	}
func InitFromFile(needToSaveSync bool, filename string) (*FileMetricsStorage, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	jsonMetrics := jsonMetrics{}
	err = json.Unmarshal(data, &jsonMetrics)
	if err != nil {
		return nil, err
	}

	storage := FileMetricsStorage{
		MemoryMetricsStorage: &MemoryMetricsStorage{metrics: jsonMetrics.Metrics},
		needToSaveSync:       needToSaveSync,
		filename:             filename,
	}
	return &storage, nil
}

// SaveToFile saves the current metrics data to the configured JSON file.
//
// The metrics are serialized as JSON with indentation for readability.
// File permissions are set to 0666 (readable and writable by all users).
//
// Returns:
//   - error: If JSON marshaling or file writing fails
//
// Example usage:
//
//	err := storage.SaveToFile()
//	if err != nil {
//	    log.Error("Failed to save metrics to file:", err)
//	}
func (s *FileMetricsStorage) SaveToFile() error {
	data, err := json.MarshalIndent(jsonMetrics{Metrics: s.metrics}, "", "   ")
	if err != nil {
		logger.Log.Error("cannot create byte data from storage", err)
		return err
	}
	err = os.WriteFile(s.filename, data, 0666)
	if err != nil {
		logger.Log.Error("cannot save to file", err)
		return err
	}
	return os.WriteFile(s.filename, data, 0666)
}

// Save stores a metric in the storage and optionally persists to file.
//
// If needToSaveSync is true, the metrics are immediately saved to file.
// If false, metrics are only stored in memory until SaveToFile() is called.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - metric: Pointer to the Metric to be saved
//
// Returns:
//   - error: If saving to memory or file fails, or if context is cancelled
//
// The method first saves the metric to memory storage, then conditionally
// persists to file based on the needToSaveSync configuration.
//
// Example usage:
//
//	err := storage.Save(ctx, &metric)
//	if err != nil {
//	    return fmt.Errorf("failed to save metric: %w", err)
//	}
func (s *FileMetricsStorage) Save(ctx context.Context, metric *models.Metric) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		err := s.MemoryMetricsStorage.Save(ctx, metric)
		if err != nil {
			return err
		}
		if s.needToSaveSync {
			err := s.SaveToFile()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
