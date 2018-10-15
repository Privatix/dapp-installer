package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dappctrl/util/log"
)

type updateCmd struct {
	name          string
	rollbackFuncs []func(conf *config, logger log.Logger)
}

func getUpdateCmd() *updateCmd {
	return &updateCmd{name: "update"}
}

func updateHelpMessage() {
	fmt.Printf(`
Usage:
	dapp-installer update [flags]

Flags:
	--config	Configuration file
	--workdir	Dapp install directory
	--source	Dapp install source
	--help		Display help information
`)
}

func (cmd *updateCmd) execute(conf *config, log log.Logger) error {
	logger := log.Add("command", cmd.name)
	if commandProcessedFlags(updateHelpMessage, conf, logger) {
		return nil
	}

	logger.Info("start process")
	defer logger.Info("finish process")

	volume := filepath.VolumeName(conf.Dapp.InstallPath)
	conf.Dapp.TempPath = util.TempPath(volume)
	defer os.RemoveAll(conf.Dapp.TempPath)

	return updateDapp(cmd, conf, logger)
}

func updateDapp(cmd *updateCmd, conf *config, logger log.Logger) error {
	oldDapp, ok := existingDapp(conf.Dapp.InstallPath, logger)

	if !ok {
		fmt.Println("dapp is not installed.")
		fmt.Println("to install run 'install' command")
		logger.Warn("dapp is not installed")
		return nil
	}

	b, dir := filepath.Split(oldDapp.InstallPath)
	newPath := filepath.Join(b, dir+"_new")
	conf.Dapp.InstallPath = newPath

	newDapp, err := initDapp(conf)
	if err != nil {
		return err
	}

	defer os.RemoveAll(newPath)

	version := util.ParseVersion(newDapp.Version)
	if util.ParseVersion(oldDapp.Version) >= version {
		fmt.Printf("dapp current version: %s, update is not required\n",
			oldDapp.Version)
		logger.Warn("already updated")
		return nil
	}

	// Update dapp core.
	if err := newDapp.Update(oldDapp, logger); err != nil {

		oldDapp.DBEngine.Start(oldDapp.InstallPath)
		oldDapp.Controller.Service.Start()

		logger.Warn(fmt.Sprintf("failed to update dapp: %v", err))
		return err
	}
	cmd.addRollbackFuncs(uninstallDapp)

	return nil
}

func (cmd *updateCmd) addRollbackFuncs(f func(c *config, l log.Logger)) {
	cmd.rollbackFuncs = append(cmd.rollbackFuncs, f)
}

func (cmd *updateCmd) rollback(conf *config, logger log.Logger) {
	for i := len(cmd.rollbackFuncs) - 1; i >= 0; i-- {
		cmd.rollbackFuncs[i](conf, logger)
	}
}
