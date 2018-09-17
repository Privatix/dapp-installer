package command

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/privatix/dapp-installer/data"
	"github.com/privatix/dapp-installer/dbengine"
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

func installProcessedFlags(cmd *installCmd, conf *config, logger log.Logger) bool {
	h := flag.Bool("help", false, "Display dapp-installer help")
	configFile := flag.String("config", "", "Configuration file")

	flag.CommandLine.Parse(os.Args[2:])

	if *h {
		cmd.helpMessage()
		return true
	}

	if *configFile == "" {
		logger.Warn("config parameter is empty")
		fmt.Println("config parameter is empty")
		return true
	}
	if err := dapputil.ReadJSONFile(*configFile, &conf); err != nil {
		logger.Error(fmt.Sprintf("failed to read config file - %s", err))
		return true
	}
	return false
}

func (cmd *installCmd) execute(conf *config, log log.Logger) error {
	logger := log.Add("command", cmd.name)
	if installProcessedFlags(cmd, conf, logger) {
		return nil
	}

	logger.Info("start process")
	defer logger.Info("finish process")

	volume := filepath.VolumeName(conf.InstallPath)
	if !util.CheckSystemPrerequisites(volume, logger) {
		logger.Warn("installation process was interrupted")
		return nil
	}
	logger.Info("check the system prerequisites was successful")

	if err := checkDatabase(cmd, conf.DBEngine, logger); err != nil {
		return err
	}

	if err := checkDapp(cmd, conf, logger); err != nil {
		return err
	}

	return finalize(cmd, conf, logger)
}

func finalize(cmd *installCmd, conf *config, logger log.Logger) error {
	if err := createRegistryKey(conf); err != nil {
		logger.Warn(fmt.Sprintf(
			"ocurred error when create registry key %v", err))
		return err
	}
	cmd.addRollbackFuncs(removeRegistry)
	logger.Info("registry keys successfully created")

	// removes backuped version of dapp core
	if conf.Dapp.BackupPath != "" {
		os.RemoveAll(conf.Dapp.BackupPath)
		logger.Info("backup version was removed")
	}
	return nil
}

func checkDapp(cmd *installCmd, conf *config, logger log.Logger) error {
	existDapp, ok := existingDapp(conf.Dapp.UserRole, logger)

	newDapp := conf.Dapp
	if ok {
		newDapp.InstallPath = existDapp.InstallPath
	}

	path := conf.Dapp.DownloadDappCtrl(conf.InstallPath)

	newDapp.Version, newDapp.InstallPath = util.DappCtrlVersion(path),
		conf.InstallPath

	if !strings.HasSuffix(newDapp.InstallPath, "\\") {
		newDapp.InstallPath += "\\"
	}

	if ok && existDapp.Version >= newDapp.Version {
		util.RemoveFile(path)
		fmt.Printf("you have not to update. your dapp version: %v\n",
			existDapp.Version)
		logger.Warn("have not to update")
		return nil
	}

	// rename existing folder name to backup
	if ok {
		existDapp.Service.Stop()
		existDapp.Service.Remove()
		newDapp.BackupPath = util.RenamePath(existDapp.InstallPath, "backup")
		if err := os.Rename(existDapp.InstallPath, newDapp.BackupPath); err != nil {
			util.RemoveFile(path)
			return err
		}
		logger.Info("existing dapp version successfully backuped")
	}

	// install dapp core
	if err := newDapp.Install(conf.DBEngine.DB, logger, ok); err != nil {
		logger.Warn(fmt.Sprintf("ocurred error when install dapp %v", err))
		return err
	}
	cmd.addRollbackFuncs(uninstallDapp)

	// execute migration scripts
	if err := dbengine.DatabaseMigrate(newDapp.InstallPath+newDapp.Controller,
		conf.DBEngine); err != nil {
		return err
	}
	cmd.addRollbackFuncs(revertDatabaseMigrate)
	logger.Info("db migration was successfully executed")

	return nil
}

func (cmd *installCmd) addRollbackFuncs(f func(c *config, l log.Logger)) {
	cmd.rollbackFuncs = append(cmd.rollbackFuncs, f)
}

// TODO(ubozov) implementation down database migration in dappctrl
func revertDatabaseMigrate(conf *config, logger log.Logger) {
}

func checkDatabase(cmd *installCmd, engine *dbengine.DBEngine,
	logger log.Logger) error {
	p, ok := util.ExistingDBEnginePort(logger)
	if !ok {
		logger.Warn("db engine is not exists")
		if err := engine.Install(logger); err != nil {
			logger.Warn(
				fmt.Sprintf("ocurred error while installing dbengine %v", err))
			return err
		}
		p, _ = util.ExistingDBEnginePort(logger)
	}
	logger.Info("checking the dbengine exists was successful")
	// update the default port number
	// by the correct port number of the existing db engine
	engine.DB.Port = strconv.Itoa(p)

	if !data.DBExists(engine.DB, logger) {
		// create dapp database
		if err := createDatabase(engine.DB); err != nil {
			logger.Warn(fmt.Sprintf(
				"ocurred error when create database %v", err))
			return err
		}
		engine.DBCreated = true
		logger.Info("database successfully created")
		cmd.addRollbackFuncs(dropDatabase)
	}
	return nil
}

func (cmd *installCmd) rollback(conf *config, logger log.Logger) {
	for i := len(cmd.rollbackFuncs) - 1; i >= 0; i-- {
		cmd.rollbackFuncs[i](conf, logger)
	}
}

func (cmd *installCmd) helpMessage() {
	fmt.Printf(`
Usage:
	dapp-installer %s [flags]

Flags:
	--config	Configuration file
	--help      Display help information
`, cmd.name)
}
