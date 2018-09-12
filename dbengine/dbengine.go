package dbengine

import (
	"fmt"
	"strings"
	"time"

	"github.com/privatix/dapp-installer/data"
	"github.com/privatix/dapp-installer/util"
)

// DBEngine has a db engine configuration.
type DBEngine struct {
	Download    string
	ServiceName string
	DataDir     string
	InstallDir  string
	Copy        []Copy
	DB          *data.DB
	DBCreated   bool
}

// Copy has a file copies parameters.
type Copy struct {
	From string
	To   string
}

// NewConfig creates a default DBEngine configuration.
func NewConfig() *DBEngine {
	return &DBEngine{
		Download:    "https://get.enterprisedb.com/postgresql/postgresql-10.5-1-windows-x64.exe",
		ServiceName: "postrgesql-10",
		DB:          data.NewConfig(),
	}
}

func (engine *DBEngine) generateInstallParams() []string {
	args := []string{"--mode", "unattended", "--unattendedmodeui", "none"}

	if len(engine.ServiceName) > 0 {
		args = append(args, "--servicename", engine.ServiceName)
	}
	if len(engine.DB.User) > 0 {
		args = append(args, "--superaccount", engine.DB.User)
	}
	if len(engine.DB.Password) > 0 {
		args = append(args, "--superpassword", engine.DB.Password)
	}
	if len(engine.InstallDir) > 0 {
		args = append(args, "--prefix", engine.InstallDir)
	}
	if len(engine.DataDir) > 0 {
		args = append(args, "--datadir", engine.DataDir)
	}
	return args
}

func interactiveWorker(s string, quit chan bool) {
	i := 0
	for {
		select {
		case <-quit:
			return
		default:
			i++
			fmt.Printf("\r%s", strings.Repeat(" ", len(s)+15))
			fmt.Printf("\r%s%s", s, strings.Repeat(".", i))
			if i >= 10 {
				i = 0
			}
			time.Sleep(time.Millisecond * 250)
		}
	}
}

// DatabaseMigrate executes migration scripts and init data.
func DatabaseMigrate(fileName string, dbEngine *DBEngine) error {
	db := dbEngine.DB
	conn := data.GetConnectionString(db.DBName, db.User, db.Password, db.Port)

	args := []string{"db-migrate", "-conn", conn}
	if err := util.ExecuteCommand(fileName, args); err != nil {
		return err
	}

	if dbEngine.DBCreated {
		args = []string{"db-init-data", "-conn", conn}
		return util.ExecuteCommand(fileName, args)
	}
	return nil
}
