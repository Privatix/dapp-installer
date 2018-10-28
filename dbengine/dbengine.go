package dbengine

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
	if err := data.CreateDatabase(engine.DB); err != nil {
		return err
	}

	if err := data.ConfigurateDatabase(engine.DB); err != nil {
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

func (engine DBEngine) databaseMigrate(fileName string) error {
	conn := engine.DB.ConnectionString()

	args := []string{"db-migrate", "-conn", conn}
	return util.ExecuteCommand(fileName, args)
}

func (engine DBEngine) databaseInit(fileName string) error {
	conn := engine.DB.ConnectionString()

	args := []string{"db-init-data", "-conn", conn}
	return util.ExecuteCommand(fileName, args)
}

// Install installs a DB engine.
func (engine *DBEngine) Install(installPath string) error {
	// install db engine
	ch := make(chan bool)
	defer close(ch)
	go util.InteractiveWorker("installation db engine", ch)

	// installs run-time components (the Visual C++ Redistributable Packages
	// for VS 2013) that are required to run postgresql database engine.
	vcredist := filepath.Join(installPath, "util/vcredist_x64.exe")
	args := []string{"/install", "/quiet", "/norestart"}
	if err := util.ExecuteCommand(vcredist, args); err != nil {
		ch <- true
		return err
	}

	// init db
	dataPath := filepath.Join(installPath, `pgsql/data`)
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		os.MkdirAll(dataPath, util.FullPermission)
	}

	util.GrantAccess(installPath)

	fileName := filepath.Join(installPath, `pgsql/bin/initdb`)
	cmd := exec.Command(fileName, "-E UTF8", "-D", dataPath)

	if err := cmd.Run(); err != nil {
		ch <- true
		return err
	}

	engine.DB.Port, _ = util.FreePort(engine.DB.Host, engine.DB.Port)

	pgconf := filepath.Join(dataPath, "postgresql.conf")
	err := configDBEngine(pgconf, engine.DB.Port)
	if err != nil {
		ch <- true
		return err
	}

	// start service
	err = engine.Start(installPath)
	if err != nil {
		ch <- true
		return err
	}

	fileName = filepath.Join(installPath, "pgsql/bin/createuser")
	if err := exec.Command(fileName, "-p", engine.DB.Port,
		"-s", engine.DB.User).Run(); err != nil {
		ch <- true
		return err
	}

	ch <- true
	fmt.Printf("\r%s\n", "db engine was successfully installed")

	return nil
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

// Hash returns db engine service unique ID.
func Hash(installPath string) string {
	hash := util.Hash(installPath)
	return fmt.Sprintf("dapp_db_%s", hash)
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
		return errors.New("failed to stopped services. timeout expired")
	}
}
