package command

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/Privatix/dapp-installer/data"
	"github.com/Privatix/dapp-installer/util"
	dapputil "github.com/Privatix/dappctrl/util"
	"github.com/Privatix/dappctrl/util/log"
)

func getInstallCmd() *Command {
	return &Command{
		Name:    "install",
		execute: install,
	}
}

func installProcessedFlags(conf *config, logger log.Logger) bool {
	h := flag.Bool("help", false, "Display dapp-installer help")
	configFile := flag.String("config", "", "Configuration file")

	flag.CommandLine.Parse(os.Args[2:])

	if *h {
		installHelp()
		return true
	}

	if *configFile != "" {
		if err := dapputil.ReadJSONFile(*configFile, &conf); err != nil {
			logger.Error(fmt.Sprintf("failed to read config file - %s", err))
			return true
		}
	}
	return false
}

func install(conf *config, logger log.Logger) {
	if installProcessedFlags(conf, logger) {
		return
	}

	logger.Info("start install process")
	defer logger.Info("finish install process")

	fmt.Println("I will be running installation process")

	if !util.CheckSystemPrerequisites(logger) {
		logger.Warn("installation process was interrupted")
		return
	}
	logger.Info("check the system prerequisites was successful")

	p, ok := util.ExistingDBEnginePort(logger)
	if !ok {
		logger.Warn("db engine is not exists")
		if err := util.InstallDBEngine(conf.DBEngine, logger); err != nil {
			logger.Warn(
				fmt.Sprintf("ocurred error while installing dbengine %v", err))
			return
		}
		p, _ = util.ExistingDBEnginePort(logger)
	}

	// update the default port number
	// by the correct port number of the existing db engine
	conf.DBEngine.DB.Port = strconv.Itoa(p)

	if !data.DBExists(conf.DBEngine.DB, logger) {
		// create dapp database
		if err := createDatabase(conf); err != nil {
			logger.Warn(fmt.Sprintf(
				"ocurred error when create database %v", err))
			return
		}
		logger.Info("database successfully created")
	}
	logger.Info("checking the dbengine exists was successful")
}

func createDatabase(conf *config) error {
	if err := data.CreateDatabase(conf.DBEngine.DB); err != nil {
		return err
	}
	if err := data.ConfigurateDatabase(conf.DBEngine.DB); err != nil {
		return err
	}
	conn := data.GetConnectionString(conf.DBEngine.DB.DBName,
		conf.DBEngine.DB.User, conf.DBEngine.DB.Password,
		conf.DBEngine.DB.Port)

	args := []string{"db-migrate", "-conn", conn}
	if err := util.ExecuteCommand(conf.DappCtrl.File, args); err != nil {
		return err
	}

	args = []string{"db-init-data", "-conn", conn}
	return util.ExecuteCommand(conf.DappCtrl.File, args)
}

func installHelp() {
	fmt.Print(`
Usage:
	dapp-installer install [flags]

Flags:
	--help      Display help information`)
}
