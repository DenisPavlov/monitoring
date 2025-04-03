package database

import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func InitDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		return nil, err
	}
	return db, nil
}
