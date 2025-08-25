package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/DenisPavlov/monitoring/internal/models"
	"github.com/DenisPavlov/monitoring/internal/util"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// attempts defines the maximum number of retry attempts for database operations
// when transient connection errors occur.
const attempts = 4

// PostgresMetricsStorage implements MetricsStorage interface for PostgreSQL database.
// It provides persistent storage for metrics with automatic retry logic for
// transient connection failures.
type PostgresMetricsStorage struct {
	db *sql.DB
}

// NewPostgresStorage creates a new PostgreSQL storage instance.
//
// Parameters:
//   - db: Initialized SQL database connection
//
// Returns:
//   - *PostgresMetricsStorage: New PostgreSQL storage instance
//   - error: If connection validation fails
//
// Example usage:
//
//	db, _ := sql.Open("pgx", "postgres://user:pass@localhost/db")
//	storage, err := NewPostgresStorage(db)
func NewPostgresStorage(db *sql.DB) (*PostgresMetricsStorage, error) {
	return &PostgresMetricsStorage{db: db}, nil
}

// InitSchema initializes the database schema by creating the metrics table
// if it doesn't already exist.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//
// Returns:
//   - error: If table creation fails
//
// The created table structure:
//   - id: TEXT PRIMARY KEY (metric identifier)
//   - type: TEXT (metric type: "gauge" or "counter")
//   - delta: BIGINT (counter value, nullable)
//   - value: DOUBLE PRECISION (gauge value, nullable)
func (s *PostgresMetricsStorage) InitSchema(ctx context.Context) error {
	return execWithRetries(func() error {
		_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS metrics (
		    id TEXT PRIMARY KEY,
		    type TEXT,
		    delta BIGINT,
		    value DOUBLE PRECISION)`,
		)
		if err != nil {
			return err
		}
		return nil
	})
}

// Save stores a single metric in the PostgreSQL database with retry logic.
// Uses appropriate UPSERT operations based on metric type.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - metric: Pointer to the Metric to be saved
//
// Returns:
//   - error: If save operation fails after all retry attempts
//
// Behavior:
//   - For gauge metrics: Replaces existing value (INSERT ON CONFLICT UPDATE)
//   - For counter metrics: Increments existing value (INSERT ON CONFLICT UPDATE with delta addition)
//   - Automatic retry on transient connection errors
func (s *PostgresMetricsStorage) Save(ctx context.Context, metric *models.Metric) error {
	return execWithRetries(func() error {
		tx, err := s.db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		err = saveOne(ctx, metric, tx)
		if err != nil {
			return err
		}
		return tx.Commit()
	})
}

// saveOne is an internal helper function that saves a single metric within a transaction.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - metric: Pointer to the Metric to be saved
//   - tx: SQL transaction object
//
// Returns:
//   - error: If database operation fails
func saveOne(ctx context.Context, metric *models.Metric, tx *sql.Tx) error {
	switch metric.MType {
	case models.GaugeMetricName:
		_, err := tx.ExecContext(ctx, `
			INSERT INTO metrics (id, type, value) VALUES ($1, $2, $3) 
			ON CONFLICT (id) DO UPDATE SET value = $3
			`, metric.ID, metric.MType, metric.Value)
		if err != nil {
			return err
		}
	case models.CounterMetricName:
		_, err := tx.ExecContext(ctx, `
			INSERT INTO metrics (id, type, delta)
			VALUES ($1, $2, $3)
			ON CONFLICT (id) DO UPDATE
			SET delta = metrics.delta + EXCLUDED.delta
		`, metric.ID, metric.MType, metric.Delta)
		if err != nil {
			return err
		}
	}
	return nil
}

// SaveAll stores multiple metrics in a single transaction with retry logic.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - metrics: Slice of Metric objects to be saved
//
// Returns:
//   - error: If any save operation fails after all retry attempts
//
// The operation is atomic - either all metrics are saved or none are.
func (s *PostgresMetricsStorage) SaveAll(ctx context.Context, metrics []models.Metric) error {
	return execWithRetries(func() error {
		tx, err := s.db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		for _, metric := range metrics {
			err := saveOne(ctx, &metric, tx)
			if err != nil {
				return err
			}
		}
		return tx.Commit()
	})
}

// GetByTypeAndID retrieves a specific metric by its ID and type with retry logic.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - ID: Metric identifier name
//   - mType: Metric type ("gauge" or "counter")
//
// Returns:
//   - models.Metric: Found metric or empty Metric if not found
//   - error: If database operation fails after all retry attempts
//
// Note: Returns empty Metric without error if no matching record is found.
func (s *PostgresMetricsStorage) GetByTypeAndID(ctx context.Context, ID, mType string) (metric models.Metric, err error) {
	return queryWithRetries(func() (models.Metric, error) {
		row := s.db.QueryRowContext(ctx, `SELECT id, type, delta, value FROM metrics WHERE id = $1 AND type = $2`, ID, mType)
		if err = row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return metric, nil
			}
			return metric, err
		}
		return metric, nil
	})
}

// GetAllByType retrieves all metrics of a specific type with retry logic.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - mType: Metric type to filter by ("gauge" or "counter")
//
// Returns:
//   - []models.Metric: Slice of metrics matching the type
//   - error: If database operation fails after all retry attempts
//
// Uses COALESCE to ensure non-null values for delta and value fields.
func (s *PostgresMetricsStorage) GetAllByType(ctx context.Context, mType string) ([]models.Metric, error) {
	return queryWithRetries(func() ([]models.Metric, error) {
		rows, err := s.db.QueryContext(ctx, `SELECT id, type, COALESCE(delta, 0), COALESCE(value,0) FROM metrics WHERE type = $1`, mType)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		metrics := make([]models.Metric, 0)
		for rows.Next() {
			var metric = models.Metric{
				Value: new(float64),
				Delta: new(int64),
			}
			if err := rows.Scan(&metric.ID, &metric.MType, metric.Delta, metric.Value); err != nil {
				return nil, err
			}
			metrics = append(metrics, metric)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return metrics, nil
	})
}

// shouldRetry determines whether a database error should trigger a retry.
// Returns true for transient PostgreSQL connection errors.
//
// Parameters:
//   - err: Error to evaluate
//
// Returns:
//   - bool: True if the operation should be retried
func shouldRetry(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.ConnectionException ||
			pgErr.Code == pgerrcode.ConnectionDoesNotExist ||
			pgErr.Code == pgerrcode.ConnectionFailure ||
			pgErr.Code == pgerrcode.SQLClientUnableToEstablishSQLConnection {
			return true
		}
	}

	return false
}

// queryWithRetries executes a database query function with automatic retry logic
// for transient connection errors.
//
// Parameters:
//   - action: Function that performs the database query and returns result
//
// Returns:
//   - T: Query result
//   - error: Final error after all retry attempts
//
// Uses exponential backoff between retry attempts.
func queryWithRetries[T any](action func() (T, error)) (result T, err error) {
	for i := 0; i < attempts; i++ {
		result, err := action()
		if shouldRetry(err) {
			time.Sleep(util.Backoff(i))
		} else {
			return result, err
		}
	}
	return result, err
}

// execWithRetries executes a database execution function with automatic retry logic
// for transient connection errors.
//
// Parameters:
//   - action: Function that performs the database operation
//
// Returns:
//   - error: Final error after all retry attempts
//
// Uses exponential backoff between retry attempts.
func execWithRetries(action func() error) (err error) {
	for i := 0; i < attempts; i++ {
		err := action()
		if shouldRetry(err) {
			time.Sleep(util.Backoff(i))
		} else {
			return err
		}
	}
	return err
}
