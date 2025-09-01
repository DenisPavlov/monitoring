// Package database provides database initialization and connection management
// functionality for PostgreSQL databases using the pgx driver.
package database

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// InitDB initializes and returns a connection to a PostgreSQL database.
//
// This function:
//   - Opens a database connection using the pgx driver
//   - Uses the standard database/sql interface
//   - Does not verify the connection with a ping (caller should handle verification)
//
// Parameters:
//   - dataSourceName: PostgreSQL connection string in the format:
//     "postgres://username:password@host:port/database?sslmode=disable"
//     or traditional connection string format
//
// Returns:
//   - *sql.DB: database connection object that can be used for queries and transactions
//   - error: if the connection cannot be established
//
// The function uses the pgx driver which is registered anonymously via the blank import.
// The returned sql.DB object represents a pool of database connections that is safe for
// concurrent use by multiple goroutines.
//
// Example usage:
//
//	db, err := database.InitDB("postgres://user:pass@localhost:5432/mydb")
//	if err != nil {
//	    log.Fatal("Failed to initialize database:", err)
//	}
//	defer db.Close()
//
//	// Verify connection with a ping
//	if err := db.Ping(); err != nil {
//	    log.Fatal("Database connection failed:", err)
//	}
//
// Note: The caller is responsible for closing the database connection
// using db.Close() when it's no longer needed.
func InitDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		return nil, err
	}
	return db, nil
}
