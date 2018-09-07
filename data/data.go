package data

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/Privatix/dappctrl/util/log"
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

// ping tests connection to database.
func ping(connStr string) error {
	conn, err := sql.Open("postgres", connStr)
	if err == nil {
		defer conn.Close()
		err = conn.Ping()
	}
	return err
}

// DBExists is checking to dappctrl database exists.
func DBExists(conf *DB, logger log.Logger) bool {
	// check to access db engine service
	connStr := GetConnectionString("postgres", conf.User, conf.Password,
		conf.Port)
	if err := ping(connStr); err != nil {
		logger.Warn(fmt.Sprintf(
			"ocurred error when check to access dbengine service %v", err))
		return false
	}

	dappConnStr := GetConnectionString(conf.DBName, conf.User,
		conf.Password, conf.Port)
	// check to access dapp database
	if err := ping(dappConnStr); err != nil {
		return false
	}
	return true
}

// CreateDatabase creates a new database.
func CreateDatabase(conf *DB) error {
	file, err := readStatikFile("/scripts/create_database.sql")
	if err != nil {
		return err
	}
	connStr := GetConnectionString("postgres", conf.User, conf.Password,
		conf.Port)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer conn.Close()

	queries := strings.Split(string(file), ";")

	for _, q := range queries {
		q = strings.Replace(q, "dappctrl", conf.DBName, -1)
		if _, err := conn.Exec(q); err != nil {
			return err
		}
	}

	return nil
}

// ConfigurateDatabase does configurate new database.
func ConfigurateDatabase(conf *DB) error {
	file, err := readStatikFile("/scripts/config_database.sql")
	if err != nil {
		return err
	}

	connStr := GetConnectionString(conf.DBName, conf.User, conf.Password,
		conf.Port)

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

// GetConnectionString is generate connection string.
func GetConnectionString(db, user, pwd, port string) string {
	connStr := "host=localhost sslmode=disable"
	return fmt.Sprintf("%s dbname=%s user=%s password=%s port=%s",
		connStr, db, user, pwd, port)
}
