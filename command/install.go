package command

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/privatix/dapp-installer/data"
	"github.com/privatix/dapp-installer/util"
	dapputil "github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

type installCmd struct {
	name          string
	rollbackFuncs []func(conf *config, logger log.Logger)
}

func getInstallCmd() *installCmd {
	return &installCmd{name: "install"}
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

func (cmd *installCmd) execute(conf *config, log log.Logger) error {
	logger := log.Add("command", cmd.name)
	if installProcessedFlags(conf, logger) {
		return nil
	}

	logger.Info("start install process")
	defer logger.Info("finish install process")

	if !util.CheckSystemPrerequisites(logger) {
		logger.Warn("installation process was interrupted")
		return nil
	}
	logger.Info("check the system prerequisites was successful")

	if err := checkDatabase(cmd, conf, logger); err != nil {
		return err
	}

	logger.Info("checkController")
	if err := checkController(cmd, conf, logger); err != nil {
		return err
	}

	return nil
}

func checkController(cmd *installCmd, conf *config, logger log.Logger) error {
	newVer := conf.DappCtrl.Version
	existVer, ok := util.ExistingDappCtrlVersion(logger)
	if !ok || newVer > existVer {
		if err := util.InstallDappCtrl(conf.InstallPath, conf.DappCtrl, logger, ok); err != nil {
			logger.Warn(fmt.Sprintf(
				"ocurred error when install dappctrl %v", err))
			return nil
		}
		if err := util.CreateRegistryKey(conf.Registry); err != nil {
			logger.Warn(fmt.Sprintf(
				"ocurred error when create registry key %v", err))
			return nil
		}
		cmd.rollbackFuncs = append(cmd.rollbackFuncs, removeRegistry)
	}
	return nil
}

func removeRegistry(conf *config, logger log.Logger) {
	if err := util.RemoveRegistryKey(conf.Registry); err != nil {
		logger.Warn(fmt.Sprintf(
			"ocurred error when remove registry key %v", err))
		return
	}
	logger.Info("windows registry successfully cleared")
}

func checkDatabase(cmd *installCmd, conf *config, logger log.Logger) error {
	p, ok := util.ExistingDBEnginePort(logger)
	if !ok {
		logger.Warn("db engine is not exists")
		if err := util.InstallDBEngine(conf.DBEngine, logger); err != nil {
			logger.Warn(
				fmt.Sprintf("ocurred error while installing dbengine %v", err))
			return err
		}
		p, _ = util.ExistingDBEnginePort(logger)
	}
	logger.Info("checking the dbengine exists was successful")
	// update the default port number
	// by the correct port number of the existing db engine
	conf.DBEngine.DB.Port = strconv.Itoa(p)

	if !data.DBExists(conf.DBEngine.DB, logger) {
		// create dapp database
		if err := createDatabase(conf); err != nil {
			logger.Warn(fmt.Sprintf(
				"ocurred error when create database %v", err))
			return err
		}
		logger.Info("database successfully created")
		cmd.rollbackFuncs = append(cmd.rollbackFuncs, dropDatabase)
	}
	return nil
}

func (cmd *installCmd) rollback(conf *config, logger log.Logger) {
	for i := len(cmd.rollbackFuncs) - 1; i >= 0; i-- {
		cmd.rollbackFuncs[i](conf, logger)
	}
}

func dropDatabase(conf *config, logger log.Logger) {
	if err := data.DropDatabase(conf.DBEngine.DB); err != nil {
		logger.Warn(fmt.Sprintf(
			"ocurred error when drop database %v", err))
		return
	}
	logger.Info("database successfully dropped")
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
