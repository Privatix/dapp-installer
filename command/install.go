package command

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/privatix/dapp-installer/dapp"
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

	conf.TempPath = util.TempPath(volume)
	defer os.RemoveAll(conf.TempPath)

	if !util.CheckSystemPrerequisites(volume, logger) {
		logger.Warn("installation process was interrupted")
		return nil
	}
	logger.Info("check the system prerequisites was successful")

	err := checkDatabase(cmd, conf.DBEngine, logger, conf.TempPath)
	if err != nil {
		return err
	}

	ok, err := checkDapp(cmd, conf, logger)
	if err != nil {
		return err
	}

	if ok {
		return finalize(cmd, conf, logger)
	}
	return nil
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

func checkDapp(cmd *installCmd, conf *config, logger log.Logger) (bool, error) {
	existDapp, ok := existingDapp(conf.Dapp.UserRole, logger)

	newDapp := initDapp(conf)
	if ok {
		newDapp.InstallPath = existDapp.InstallPath
	}

	version := util.ParseVersion(newDapp.Version)
	if ok && util.ParseVersion(existDapp.Version) >= version {
		fmt.Printf("you don't need to update. your dapp version: %v\n",
			existDapp.Version)
		logger.Warn("don't need to update")
		return false, nil
	}

	// rename existing folder name to backup
	if ok {
		existDapp.Service.Uninstall()
		newDapp.BackupPath = util.RenamePath(existDapp.InstallPath, "backup")
		if err := os.Rename(existDapp.InstallPath, newDapp.BackupPath); err != nil {
			return false, err
		}
		logger.Info("existing dapp version successfully backuped")
	}

	// execute migration scripts
	if err := dbengine.DatabaseMigrate(newDapp.TempPath+newDapp.Controller,
		conf.DBEngine); err != nil {
		return false, err
	}
	cmd.addRollbackFuncs(revertDatabaseMigrate)
	logger.Info("db migration was successfully executed")

	// install dapp core
	if err := newDapp.Install(conf.DBEngine.DB, logger, ok); err != nil {
		logger.Warn(fmt.Sprintf("ocurred error when install dapp %v", err))
		return false, err
	}
	cmd.addRollbackFuncs(uninstallDapp)

	return true, nil
}

func initDapp(conf *config) *dapp.Dapp {
	d := conf.Dapp
	d.TempPath = conf.TempPath
	path := conf.Dapp.DownloadDappCtrl()

	d.Version, d.InstallPath = util.DappCtrlVersion(path),
		conf.InstallPath
	_, d.Controller = filepath.Split(d.DownloadCtrl)

	if !strings.HasSuffix(d.TempPath, "\\") {
		d.TempPath += "\\"
	}
	if !strings.HasSuffix(d.InstallPath, "\\") {
		d.InstallPath += "\\"
	}
	return d
}

func (cmd *installCmd) addRollbackFuncs(f func(c *config, l log.Logger)) {
	cmd.rollbackFuncs = append(cmd.rollbackFuncs, f)
}

// TODO(ubozov) implementation down database migration in dappctrl
func revertDatabaseMigrate(conf *config, logger log.Logger) {
}

func checkDatabase(cmd *installCmd, engine *dbengine.DBEngine,
	logger log.Logger, tempPath string) error {
	p, ok := util.ExistingDBEnginePort(logger)
	if !ok {
		logger.Warn("db engine is not exists")
		if err := engine.Install(tempPath, logger); err != nil {
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
