package dbengine

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/privatix/dapp-installer/data"
	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dappctrl/util/log"
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
func (engine *DBEngine) Install(installPath string, logger log.Logger) error {
	// install db engine
	ch := make(chan bool)
	defer close(ch)
	go util.InteractiveWorker("Installation DB Engine", ch)

	// init db
	dataPath := filepath.Join(installPath, `pgsql/data`)
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		os.MkdirAll(dataPath, util.FullPermission)
	}

	u, err := user.Current()
	if err != nil {
		ch <- true
		return err
	}

	util.GrantAccess(installPath, u.Username)

	fileName := filepath.Join(installPath, `pgsql/bin/initdb`)
	cmd := exec.Command(fileName, "-E UTF8", "-D", dataPath)

	if err := cmd.Run(); err != nil {
		ch <- true
		return err
	}

	engine.DB.Port, _ = util.FreePort(engine.DB.Host, engine.DB.Port)

	pgconf := filepath.Join(dataPath, "postgresql.conf")
	err = configDBEngine(pgconf, engine.DB.Port)

	// start service
	err = startService(installPath, u.Username)
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
	fmt.Printf("\r%s\n", "DB Engine successfully installed")
	logger.Info("dbengine successfully installed")

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
func (engine *DBEngine) Remove(installPath string, logger log.Logger) error {
	return removeService(installPath)
}

// Start starts the DB engine.
func (engine *DBEngine) Start(installPath string) error {
	return startService(installPath, engine.DB.User)
}

// Stop stops the DB engine.
func (engine *DBEngine) Stop(installPath string) error {
	return stopService(installPath)
}
