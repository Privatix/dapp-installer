package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dappctrl/util/log"
)

type installCmd struct {
	name          string
	rollbackFuncs []func(conf *config, logger log.Logger)
}

func getInstallCmd() *installCmd {
	return &installCmd{name: "install"}
}

func (cmd *installCmd) execute(conf *config, log log.Logger) error {
	logger := log.Add("command", cmd.name)
	if commandProcessedFlags(installHelpMessage, conf, logger) {
		return nil
	}

	logger.Info("start process")
	defer logger.Info("finish process")

	volume := filepath.VolumeName(conf.Dapp.InstallPath)

	if !util.CheckSystemPrerequisites(volume, logger) {
		logger.Warn("installation process was interrupted")
		return nil
	}
	logger.Info("check the system prerequisites was successful")

	conf.Dapp.TempPath = util.TempPath(volume)
	defer os.RemoveAll(conf.Dapp.TempPath)

	return installDapp(cmd, conf, logger)
}

func installDapp(cmd *installCmd, conf *config, logger log.Logger) error {
	existDapp, ok := existingDapp(conf.Dapp.InstallPath, logger)

	if ok {
		fmt.Printf("dapp is installed in '%s'. your dapp version: %v\n",
			conf.Dapp.InstallPath, existDapp.Version)
		fmt.Println("to update run 'update' command")
		logger.Warn("dapp is installed")
		return nil
	}

	newDapp, err := initDapp(conf)
	if err != nil {
		return err
	}

	// install dapp core
	if err := newDapp.Install(logger); err != nil {
		newDapp.Remove(logger)
		logger.Warn(fmt.Sprintf("failed to install dapp: %v", err))
		return err
	}
	cmd.addRollbackFuncs(uninstallDapp)

	return nil
}

func (cmd *installCmd) addRollbackFuncs(f func(c *config, l log.Logger)) {
	cmd.rollbackFuncs = append(cmd.rollbackFuncs, f)
}

func (cmd *installCmd) rollback(conf *config, logger log.Logger) {
	for i := len(cmd.rollbackFuncs) - 1; i >= 0; i-- {
		cmd.rollbackFuncs[i](conf, logger)
	}
}

func installHelpMessage() {
	fmt.Printf(`
Usage:
	dapp-installer install [flags]

Flags:
	--config	Configuration file
	--role		Dapp user role
	--workdir	Dapp install directory
	--source	Dapp install source
	--help		Display help information
`)
}
