package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DenisPavlov/monitoring/internal/models"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(ctx context.Context, db *sql.DB) (*PostgresStorage, error) {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS metrics (
		    id TEXT,
		    type TEXT,
		    delta BIGINT,
		    value DOUBLE PRECISION,
		    PRIMARY KEY (id, type))`,
	)
	if err != nil {
		return nil, err
	}
	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) Save(ctx context.Context, metric *models.Metrics) error {
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
}

func saveOne(ctx context.Context, metric *models.Metrics, tx *sql.Tx) error {
	switch metric.MType {
	case models.GaugeMetricName:
		_, err := tx.ExecContext(ctx, `INSERT INTO metrics (id, type, value) VALUES ($1, $2, $3) ON CONFLICT (id, type) DO UPDATE SET value = $3`, metric.ID, metric.MType, metric.Value)
		if err != nil {
			return err
		}
	case models.CounterMetricName:
		row := tx.QueryRowContext(ctx, `SELECT delta FROM metrics WHERE id = $1 AND type = $2 FOR UPDATE`, metric.ID, metric.MType)

		var currentDelta int64
		err := row.Scan(&currentDelta)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		*metric.Delta = *metric.Delta + currentDelta
		_, err = tx.ExecContext(ctx, `INSERT INTO metrics (id, type, delta) VALUES ($1, $2, $3) ON CONFLICT (id, type) DO UPDATE SET delta = $3`, metric.ID, metric.MType, metric.Delta)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStorage) SaveAll(ctx context.Context, metrics []models.Metrics) error {
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
}

func (s *PostgresStorage) GetByTypeAndID(ctx context.Context, ID, mType string) (metric models.Metrics, err error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, type, delta, value FROM metrics WHERE id = $1 AND type = $2`, ID, mType)
	if err = row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return metric, nil
		}
		return metric, err
	}
	return metric, nil
}

func (s *PostgresStorage) GetAllByType(ctx context.Context, mType string) ([]models.Metrics, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, type, delta, value FROM metrics WHERE type = $1`, mType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metrics := make([]models.Metrics, 0)
	for rows.Next() {
		var metric models.Metrics
		if err := rows.Scan(&metric.ID, &metric.MType, metric.Delta, metric.Value); err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return metrics, nil
}
