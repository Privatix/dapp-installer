package data

import (
	"database/sql"
	"fmt"

	// Load Go Postgres driver.
	_ "github.com/lib/pq"
)

const (
	versionKey = `system.version.app`
)

// DB has a configuration for database connect.
type DB struct {
	Host     string
	User     string
	Password string
	DBName   string
	Port     string
}

// NewConfig creates a default DB configuration.
func NewConfig() *DB {
	return &DB{
		Host:   "127.0.0.1",
		DBName: "dappctrl",
		User:   "postgres",
		Port:   "5433",
	}
}

// ConnectionString returns the DB connection string.
func (db *DB) ConnectionString() string {
	return getConnectionString(db.Host, db.DBName, db.User,
		db.Password, db.Port)
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

// UpdateSetting updates a setting in db.
func UpdateSetting(db *DB, key, value string) error {
	sqlStatement := `UPDATE settings SET value = $2 WHERE key = $1;`
	return execQuery(db, sqlStatement, key, value)
}

func execQuery(db *DB, query string, args ...interface{}) error {
	conn, err := sql.Open("postgres", db.ConnectionString())
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Exec(query, args...)

	return err
}

func getConnectionString(host, db, user, pwd, port string) string {
	connStr := "sslmode=disable"
	return fmt.Sprintf("%s host=%s dbname=%s user=%s port=%s password=%s",
		connStr, host, db, user, port, pwd)
}
