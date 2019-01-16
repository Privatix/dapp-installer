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
		Host:   "localhost",
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

// WriteAppVersion writes AppVersion.
func WriteAppVersion(db *DB, version string) error {
	conn, err := sql.Open("postgres", db.ConnectionString())
	if err != nil {
		return err
	}
	defer conn.Close()

	var sqlStatement string
	if _, ok := ReadAppVersion(db); ok {
		sqlStatement = `UPDATE settings SET value = $2 WHERE key = $1;`
	} else {
		sqlStatement = `INSERT INTO
		settings (key, value, permissions, description, name)
		VALUES ($1, $2, 1, 'Version of application', 'app version');`
	}

	_, err = conn.Exec(sqlStatement, versionKey, version)

	return err
}

// ReadAppVersion returns AppVersion.
func ReadAppVersion(db *DB) (string, bool) {
	conn, err := sql.Open("postgres", db.ConnectionString())
	if err != nil {
		return "", false
	}
	defer conn.Close()

	var value string
	row := conn.QueryRow(`SELECT value FROM settings WHERE key = $1;`,
		versionKey)
	if err := row.Scan(&value); err != nil {
		return "", false
	}
	return value, true
}

func getConnectionString(host, db, user, pwd, port string) string {
	connStr := "sslmode=disable"
	return fmt.Sprintf("%s host=%s dbname=%s user=%s port=%s password=%s",
		connStr, host, db, user, port, pwd)
}
