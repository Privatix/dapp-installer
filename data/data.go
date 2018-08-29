package data

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"strings"

	// Load Go Postgres driver.
	_ "github.com/lib/pq"
	"github.com/rakyll/statik/fs"
)

//go:generate statik -p data -f -src=. -dest=..

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

// CreateDatabase creates a new database.
func CreateDatabase(dbname, connStr string) error {
	file, err := readStatikFile("/scripts/create_database.sql")
	if err != nil {
		return err
	}

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer conn.Close()

	queries := strings.Split(string(file), ";")

	for _, q := range queries {
		q = strings.Replace(q, "dappctrl", dbname, -1)
		if _, err := conn.Exec(q); err != nil {
			return err
		}
	}

	return nil
}

// ConfigurateDatabase does configurate new database.
func ConfigurateDatabase(connStr string) error {
	file, err := readStatikFile("/scripts/config_database.sql")
	if err != nil {
		return err
	}

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer conn.Close()

	queries := strings.Split(string(file), ";")

	for _, q := range queries {
		if _, err := conn.Exec(q); err != nil {
			return err
		}
	}

	return nil
}

func readStatikFile(name string) ([]byte, error) {
	fs, err := fs.New()
	if err != nil {
		return nil, errors.New("failed to open statik filesystem")
	}

	file, err := fs.Open(name)
	if err != nil {
		return nil, errors.New("failed to open statik file")
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.New("failed to read statik file")
	}

	return data, nil
}
