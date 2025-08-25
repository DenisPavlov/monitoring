package storage

import (
	"context"

	"github.com/DenisPavlov/monitoring/internal/models"
)

// MetricsStorage defines the interface for storing and retrieving metrics.
// Implementations can provide in-memory, file-based, or database storage.
//
// The interface is designed to be implemented by various storage backends:
//   - MemoryMetricsStorage: In-memory storage with map-based implementation
//   - FileMetricsStorage: File-based storage with JSON persistence
//   - PostgresMetricsStorage: PostgreSQL database storage
//
// All methods support context for cancellation and timeout propagation.
type MetricsStorage interface {
	// Save stores a single metric in the storage.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - metric: Pointer to the Metric to be saved
	//
	// Returns:
	//   - error: If the save operation fails
	//
	// Behavior varies by implementation:
	//   - For gauge metrics: Typically replaces existing value
	//   - For counter metrics: Typically increments existing value
	Save(ctx context.Context, metric *models.Metric) error

	// SaveAll stores multiple metrics in the storage atomically.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - metrics: Slice of Metric objects to be saved
	//
	// Returns:
	//   - error: If any save operation fails
	//
	// Implementations should ensure that either all metrics are saved
	// successfully or none are (atomic operation).
	SaveAll(ctx context.Context, metrics []models.Metric) error

	// GetByTypeAndID retrieves a specific metric by its identifier and type.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - ID: Metric identifier name
	//   - mType: Metric type ("gauge" or "counter")
	//
	// Returns:
	//   - models.Metric: Found metric or empty Metric if not found
	//   - error: If the retrieval operation fails
	//
	// Implementations should return an empty Metric without error
	// when the requested metric is not found.
	GetByTypeAndID(ctx context.Context, ID, mType string) (models.Metric, error)

	// GetAllByType retrieves all metrics of a specific type.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - mType: Metric type to filter by ("gauge" or "counter")
	//
	// Returns:
	//   - []models.Metric: Slice of metrics matching the type
	//   - error: If the retrieval operation fails
	//
	// Returns an empty slice if no metrics of the specified type are found.
	GetAllByType(ctx context.Context, mType string) ([]models.Metric, error)
}
