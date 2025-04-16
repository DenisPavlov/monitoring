package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DenisPavlov/monitoring/internal/models"
	"github.com/DenisPavlov/monitoring/internal/util"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

const attempts = 4

type PostgresMetricsStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) (*PostgresMetricsStorage, error) {
	return &PostgresMetricsStorage{db: db}, nil
}

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

func (s *PostgresMetricsStorage) GetAllByType(ctx context.Context, mType string) ([]models.Metric, error) {
	return queryWithRetries(func() ([]models.Metric, error) {
		rows, err := s.db.QueryContext(ctx, `SELECT id, type, delta, value FROM metrics WHERE type = $1`, mType)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		metrics := make([]models.Metric, 0)
		for rows.Next() {
			var metric models.Metric
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
