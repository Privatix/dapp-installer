package data

import (
	"database/sql"
	// Load Go Postgres driver.
	_ "github.com/lib/pq"
)

// DB has a configuration for database connect.
type DB struct {
	User     string
	Password string
	DBName   string
	Port     string
}

// Ping tests connection to database.
func Ping(connStr string) error {
	conn, err := sql.Open("postgres", connStr)
	if err == nil {
		defer conn.Close()
		err = conn.Ping()
	}
	return err
}
