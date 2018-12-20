package dbengine

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/privatix/dapp-installer/data"
	"github.com/privatix/dapp-installer/util"
)

// DBEngine has a db engine configuration.
type DBEngine struct {
	ServiceName string
	DB          *data.DB
}

// NewConfig creates a default DBEngine configuration.
func NewConfig() *DBEngine {
	return &DBEngine{
		DB: data.NewConfig(),
	}
}

// CreateDatabase creates new database.
func (engine *DBEngine) CreateDatabase(fileName string) error {
	if err := engine.createDatabase(fileName); err != nil {
		return err
	}

	if err := engine.databaseMigrate(fileName); err != nil {
		return err
	}

	return engine.databaseInit(fileName)
}

// UpdateDatabase executes db migrations scripts.
func (engine DBEngine) UpdateDatabase(fileName string) error {
	return engine.databaseMigrate(fileName)
}

func (engine DBEngine) createDatabase(fileName string) error {
	db := &data.DB{
		Host:     engine.DB.Host,
		User:     engine.DB.User,
		Password: engine.DB.Password,
		DBName:   "postgres",
		Port:     engine.DB.Port,
	}
	conn := db.ConnectionString()
	return util.ExecuteCommand(fileName, "db-create", "-conn", conn)
}

func (engine DBEngine) databaseMigrate(fileName string) error {
	conn := engine.DB.ConnectionString()
	return util.ExecuteCommand(fileName, "db-migrate", "-conn", conn)
}

func (engine DBEngine) databaseInit(fileName string) error {
	conn := engine.DB.ConnectionString()
	return util.ExecuteCommand(fileName, "db-init-data", "-conn", conn)
}

// Install installs a DB engine.
func (engine *DBEngine) Install(installPath string) error {
	if err := prepareToInstall(installPath); err != nil {
		return err
	}

	// init db
	dataPath := filepath.Join(installPath, "pgsql", "data")
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		os.MkdirAll(dataPath, util.FullPermission)
	}

	util.GrantAccess(installPath)

	fileName := filepath.Join(installPath, "pgsql", "bin", "initdb")
	err := util.ExecuteCommand(fileName, "-E UTF8", "-D", dataPath)

	if err != nil {
		return err
	}

	engine.DB.Port, _ = util.FreePort(engine.DB.Host, engine.DB.Port)

	pgconf := filepath.Join(dataPath, "postgresql.conf")
	if err := configDBEngine(pgconf, engine.DB.Port); err != nil {
		return err
	}

	// start service
	if err := engine.Start(installPath); err != nil {
		return err
	}

	fileName = filepath.Join(installPath, "pgsql", "bin", "createuser")
	args := []string{"-p", engine.DB.Port, "-s", engine.DB.User}
	return createUser(fileName, args...)
}

func createUser(fileName string, args ...string) error {
	done := make(chan bool)
	go func() {
		for {
			err := util.ExecuteCommand(fileName, args...)
			if err == nil {
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
		done <- true
	}()

	select {
	case <-done:
		return nil
	case <-time.After(util.Timeout):
		return errors.New("failed to createuser. timeout expired")
	}
}

func configDBEngine(pgconf, port string) error {
	read, err := ioutil.ReadFile(pgconf)
	if err != nil {
		return err
	}

	newContents := strings.Replace(string(read),
		"#port = 5432", "port = "+port, -1)

	return ioutil.WriteFile(pgconf, []byte(newContents), 0)
}

// Remove removes the DB engine.
func (engine *DBEngine) Remove(installPath string) error {
	return removeService(installPath)
}

// Start starts the DB engine.
func (engine *DBEngine) Start(installPath string) error {
	if err := startService(installPath); err != nil {
		return err
	}
	return engine.checkRunning()
}

// Stop stops the DB engine.
func (engine *DBEngine) Stop(installPath string) error {
	return stopService(installPath)
}

func (engine *DBEngine) checkRunning() error {
	done := make(chan bool)
	go func() {
		for {
			p, _ := util.FreePort(engine.DB.Host, engine.DB.Port)
			if p != engine.DB.Port {
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
		done <- true
	}()

	select {
	case <-done:
		return nil
	case <-time.After(util.Timeout):
		return errors.New("failed to check running dbengine. timeout expired")
	}
}
