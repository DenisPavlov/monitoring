package storage

//
//import (
//	"context"
//	"database/sql"
//	"github.com/DenisPavlov/monitoring/internal/logger"
//)
//
//type PostgresStorage struct {
//	db *sql.DB
//}
//
//func NewPostgresStorage(ctx context.Context, db *sql.DB) (*PostgresStorage, error) {
//	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS gauge (name TEXT UNIQUE NOT NULL, value DOUBLE PRECISION )`)
//	if err != nil {
//		return nil, err
//	}
//
//	_, err = db.ExecContext(ctx, `CREATE TABLE  IF NOT EXISTS counter (name TEXT UNIQUE NOT NULL, value INT )`)
//	if err != nil {
//		return nil, err
//	}
//
//	return &PostgresStorage{
//		db: db,
//	}, nil
//}
//
//func (storage *PostgresStorage) AddGauge(name string, value float64) error {
//	_, err := storage.db.Exec(`INSERT INTO gauge (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value = $2`, name, value)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (storage *PostgresStorage) Gauge(name string) (float64, bool) {
//	row := storage.db.QueryRow("SELECT value FROM gauge WHERE name = $1", name)
//	var value float64
//	if err := row.Scan(&value); err != nil {
//		return 0, false
//	}
//	return value, true
//}
//
//func (storage *PostgresStorage) AddCounter(name string, value int64) error {
//	_, err := storage.db.Exec(`INSERT INTO counter (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value = $2`, name, value)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (storage *PostgresStorage) Counter(name string) (int64, bool) {
//	row := storage.db.QueryRow("SELECT value FROM counter WHERE name = $1", name)
//	var value int64
//	if err := row.Scan(&value); err != nil {
//		return 0, false
//	}
//	return value, true
//}
//
//func (storage *PostgresStorage) AllGauges() map[string]float64 {
//	rows, err := storage.db.Query(`SELECT name, value FROM gauge`)
//	if err != nil {
//		return nil
//	}
//	defer rows.Close()
//
//	var result = make(map[string]float64)
//	for rows.Next() {
//		var name string
//		var value float64
//		if err := rows.Scan(&name, &value); err != nil {
//			logger.Log.Error("Error scanning row:", err)
//		}
//		result[name] = value
//	}
//	err = rows.Err()
//	if err != nil {
//		logger.Log.Error("Error scanning rows:", err)
//		return nil
//	}
//	return result
//}
//
//func (storage *PostgresStorage) AllCounters() map[string]int64 {
//	rows, err := storage.db.Query(`SELECT name, value FROM counter`)
//	if err != nil {
//		return nil
//	}
//	defer rows.Close()
//
//	var result = make(map[string]int64)
//	for rows.Next() {
//		var name string
//		var value int64
//		if err := rows.Scan(&name, &value); err != nil {
//			logger.Log.Error("Error scanning row:", err)
//		}
//		result[name] = value
//	}
//	err = rows.Err()
//	if err != nil {
//		logger.Log.Error("Error scanning rows:", err)
//		return nil
//	}
//	return result
//}
