package dbengine

import (
	"context"
	"errors"
	"fmt"
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
	Timeout     uint64 // in seconds
	Autostart   bool
}

// NewConfig creates a default DBEngine configuration.
func NewConfig() *DBEngine {
	return &DBEngine{
		DB:      data.NewConfig(),
		Timeout: 300,
	}
}

// CreateDatabase creates new database.
func (engine *DBEngine) CreateDatabase(fileName string) error {
	if err := engine.checkRunning(); err != nil {
		return err
	}

	if err := engine.executor(engine.createDatabase, fileName); err != nil {
		return err
	}

	if err := engine.executor(engine.databaseMigrate, fileName); err != nil {
		return err
	}

	return engine.loadproddata(fileName)
}

// UpdateDatabase executes db migrations scripts.
func (engine DBEngine) UpdateDatabase(fileName string) error {
	if err := engine.databaseMigrate(fileName); err != nil {
		return err
	}

	return engine.loadproddata(fileName)
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
	if err := engine.Ping(); err != nil {
		return err
	}
	conn := engine.DB.ConnectionString()
	return util.ExecuteCommand(fileName, "db-migrate", "-conn", conn)
}

func (engine DBEngine) loadproddata(fileName string) error {
	conn := engine.DB.ConnectionString()
	return util.ExecuteCommand(fileName, "db-load-data", "-conn", conn)
}

// Install installs a DB engine.
func (engine *DBEngine) Install(installPath, username, installUID string) error {
	if err := prepareToInstall(installPath); err != nil {
		return fmt.Errorf("failed to prepare install db: %v", err)
	}

	// init db
	dataPath := filepath.Join(installPath, "pgsql", "data")
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dataPath, util.FullPermission); err != nil {
			return fmt.Errorf("could not create data direcotry: %v", err)
		}
		if username != "" {
			err := util.ExecuteCommand("chown", "-R", username, dataPath)
			if err != nil {
				return fmt.Errorf("could not change data dir owner: %v", err)
			}
		}
	}

	if username == "" {
		util.GrantAccess(installPath)
	}

	fileName := filepath.Join(installPath, "pgsql", "bin", "initdb")
	var err error
	if username != "" {
		command := fmt.Sprint(fileName, " -D ", dataPath)
		err = util.ExecuteCommand("osascript", "-e", fmt.Sprintf(`do shell script "sudo -u %s %s"`, username, command))
	} else {
		err = util.ExecuteCommand(fileName, "-D", dataPath)
	}

	if err != nil {
		return fmt.Errorf("failed to init db: %v", err)
	}

	engine.DB.Port, _ = util.FreePort(engine.DB.Host, engine.DB.Port)

	pgconf := filepath.Join(dataPath, "postgresql.conf")
	if err := SetPort(pgconf, "5432", engine.DB.Port); err != nil {
		return fmt.Errorf("failed to configure db conf: %v", err)
	}

	// start service
	if err := engine.Start(installPath, installUID); err != nil {
		return fmt.Errorf("failed to start db engine: %v", err)
	}

	fileName = filepath.Join(installPath, "pgsql", "bin", "createuser")
	return engine.createUser(fileName, username)
}

func (engine *DBEngine) createUser(fileName, username string) error {
	args := []string{"-p", engine.DB.Port, "-s", engine.DB.User}

	done := make(chan bool)
	go func() {
		for {
			var err error
			if username == "" {
				err = util.ExecuteCommand(fileName, args...)
			} else {
				command := fmt.Sprintf("%s %s", fileName, strings.Join(args, " "))
				err = util.ExecuteCommand("osascript", "-e", fmt.Sprintf(`do shell script "sudo -u %s %s"`, username, command))
			}
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
	case <-time.After(util.TimeOutInSec(engine.Timeout)):
		return errors.New("failed to createuser. timeout expired")
	}
}

// SetPort sets db engine port number.
func SetPort(pgconf, oldPort, newPort string) error {
	read, err := ioutil.ReadFile(pgconf)
	if err != nil {
		return err
	}

	p := "port = "
	newContents := strings.Replace(string(read), p+oldPort, p+newPort, -1)
	newContents = strings.Replace(newContents, "#"+p, p, -1)

	return ioutil.WriteFile(pgconf, []byte(newContents), 0)
}

// Remove removes the DB engine.
func (engine *DBEngine) Remove(installPath, installUID string) error {
	return removeService(installPath, installUID)
}

// Start starts the DB engine.
func (engine *DBEngine) Start(installPath, installUID string) error {
	if err := startService(installPath, installUID, engine.Autostart); err != nil {
		return err
	}
	if err := engine.checkRunning(); err != nil {
		return err
	}
	// Wait for db system finish starting up.
	time.Sleep(time.Second)

	return nil
}

// Stop stops the DB engine.
func (engine *DBEngine) Stop(installPath, installUID string) error {
	return stopService(installPath, installUID)
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
	case <-time.After(util.TimeOutInSec(engine.Timeout)):
		return errors.New("failed to check running dbengine. timeout expired")
	}
}

// Ping tests connection to database.
func (engine DBEngine) Ping() error {
	conn := engine.DB.ConnectionString()

	ctx, cancel := context.WithTimeout(context.Background(),
		util.TimeOutInSec(engine.Timeout))
	defer cancel()
	return util.RetryTillSucceed(ctx, func() error { return data.Ping(conn) })
}

func (engine DBEngine) executor(f func(string) error, param string) error {
	ctx, cancel := context.WithTimeout(context.Background(),
		util.TimeOutInSec(engine.Timeout))
	defer cancel()
	return util.RetryTillSucceed(ctx, func() error { return f(param) })
}
