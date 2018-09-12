package command

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/privatix/dapp-installer/dapp"
	"github.com/privatix/dapp-installer/data"
	"github.com/privatix/dapp-installer/dbengine"
	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dappctrl/util/log"
)

const mainHelpMsg = `
Usage:
  dapp-installer [command] [flags]

Available Commands:
  install     Install dapp core
  update      Update dapp core
  remove      Remove dapp core

Flags:
  --help      Display help information
  --version   Display the current version of this CLI
  
Use "dapp-installer [command] --help" for more information about a command.
`

type command interface {
	execute(conf *config, logger log.Logger) error
	rollback(conf *config, logger log.Logger)
}

type config struct {
	InstallPath string
	DBEngine    *dbengine.DBEngine
	Registry    *util.Registry
	Dapp        *dapp.Dapp
}

func newConfig() *config {
	return &config{
		Dapp:     dapp.NewConfig(),
		DBEngine: dbengine.NewConfig(),
	}
}

// Execute has a command execute
func Execute(logger log.Logger, printVersion func(), args []string) {
	if len(args) == 0 {
		fmt.Print(mainHelpMsg)
		return
	}

	commands := map[string]command{
		"install": getInstallCmd(),
		"update":  getUpdateCmd(),
		"remove":  getRemoveCmd(),
	}

	cmd, exist := commands[strings.ToLower(args[0])]
	if !exist {
		if !processedFlags(printVersion) {
			fmt.Println("unknown command or flags:", args[0])
			fmt.Print(mainHelpMsg)
		}
		return
	}

	conf := newConfig()
	if err := cmd.execute(conf, logger); err != nil {
		cmd.rollback(conf, logger)
		fmt.Println("installation was canceled after an exception occurred.")
		logger.Info("installation was canceled")
		return
	}
	fmt.Println("command successfully finished")
}

func processedFlags(printVersion func()) bool {
	v := flag.Bool("version", false, "Prints current dapp-installer version")
	h := flag.Bool("help", false, "Display dapp-installer help")

	flag.Parse()

	if *v {
		printVersion()
		return true
	}
	if *h {
		fmt.Printf(mainHelpMsg)
		return true
	}
	return false
}

func createRegistryKey(conf *config) error {
	conf.Registry.Install = append(conf.Registry.Install, util.Key{
		Name: "BaseDirectory", Type: "string", Value: conf.Dapp.InstallPath,
	})
	conf.Registry.Install = append(conf.Registry.Install, util.Key{
		Name: "Version", Type: "string", Value: conf.Dapp.Version,
	})
	conf.Registry.Install = append(conf.Registry.Install, util.Key{
		Name: "Database", Type: "string", Value: conf.DBEngine.DB.DBName,
	})
	conf.Registry.Install = append(conf.Registry.Install, util.Key{
		Name: "ServiceID", Type: "string", Value: conf.Dapp.Service.GUID,
	})
	conf.Registry.Install = append(conf.Registry.Install, util.Key{
		Name: "Controller", Type: "string", Value: conf.Dapp.Controller,
	})

	conf.Registry.Uninstall = append(conf.Registry.Uninstall, util.Key{
		Name: "InstallLocation", Type: "string", Value: conf.Dapp.InstallPath,
	})
	current := fmt.Sprintf("%d%d%d", time.Now().Year(),
		time.Now().Month(), time.Now().Day())
	conf.Registry.Uninstall = append(conf.Registry.Uninstall, util.Key{
		Name: "InstallDate", Type: "string", Value: current,
	})
	conf.Registry.Uninstall = append(conf.Registry.Uninstall, util.Key{
		Name: "DisplayVersion", Type: "string", Value: conf.Dapp.Version,
	})
	conf.Registry.Uninstall = append(conf.Registry.Uninstall, util.Key{
		Name: "DisplayName", Type: "string",
		Value: "Privatix Dapp " + conf.Dapp.UserRole,
	})
	return util.CreateRegistryKey(conf.Registry, conf.Dapp.UserRole)
}

func removeRegistry(conf *config, logger log.Logger) {
	err := util.RemoveRegistryKey(conf.Registry, conf.Dapp.UserRole)
	if err != nil {
		logger.Warn(fmt.Sprintf(
			"ocurred error when remove registry key %v", err))
		return
	}
	logger.Info("windows registry successfully cleared")
}

func dropDatabase(conf *config, logger log.Logger) {
	if err := data.DropDatabase(conf.DBEngine.DB); err != nil {
		logger.Warn(fmt.Sprintf(
			"ocurred error when drop database %v", err))
		return
	}
	logger.Info("database successfully dropped")
}

func createDatabase(db *data.DB) error {
	if err := data.CreateDatabase(db); err != nil {
		return err
	}
	return data.ConfigurateDatabase(db)
}

func uninstallDapp(conf *config, logger log.Logger) {
	err := conf.Dapp.Remove()
	if err != nil {
		logger.Warn(fmt.Sprintf(
			"ocurred error when remove dapp %v", err))
		return
	}

	if conf.Dapp.BackupPath == "" {
		logger.Info("dappctrl successfully removed")
		return
	}

	if err := os.Rename(conf.Dapp.BackupPath, conf.Dapp.InstallPath); err != nil {
		logger.Warn(fmt.Sprintf(
			"ocurred error when restore dapp %v", err))
		return
	}
	conf.Dapp.Service.Install()
	conf.Dapp.Service.Start()

	logger.Info("dappctrl successfully removed")
}
