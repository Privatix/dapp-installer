package data

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/privatix/dapp-installer/statik"
	"github.com/privatix/dappctrl/util/log"

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

// NewConfig creates a default DB configuration.
func NewConfig() *DB {
	return &DB{
		DBName:   "dappctrl",
		User:     "postgres",
		Password: "postgres",
		Port:     "5432",
	}
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
	file, err := statik.ReadFile("/scripts/create_database.sql")
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

// DropDatabase removes the database.
func DropDatabase(conf *DB) error {
	connStr := GetConnectionString("postgres", conf.User, conf.Password,
		conf.Port)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer conn.Close()
	query := fmt.Sprintf("DROP DATABASE %s;", conf.DBName)
	if _, err := conn.Exec(query); err != nil {
		return err
	}

	return nil
}

// ConfigurateDatabase does configurate new database.
func ConfigurateDatabase(conf *DB) error {
	file, err := statik.ReadFile("/scripts/config_database.sql")
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

// GetConnectionString is generate connection string.
func GetConnectionString(db, user, pwd, port string) string {
	connStr := "host=localhost sslmode=disable"
	return fmt.Sprintf("%s dbname=%s user=%s password=%s port=%s",
		connStr, db, user, pwd, port)
}
