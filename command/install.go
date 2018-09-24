package command

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

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
	existDapp, ok := existingDapp(conf.Dapp.UserRole, logger)

	if ok {
		fmt.Printf("dapp was already installed. your dapp version: %v\n",
			existDapp.Version)
		fmt.Println("for update you should run with command 'update'")
		logger.Warn("dapp was already installed")
		return nil
	}

	newDapp, err := initDapp(conf)
	if err != nil {
		return err
	}

	// install dapp core
	if err := newDapp.Install(logger); err != nil {
		newDapp.Remove(logger)
		logger.Warn(fmt.Sprintf("ocurred error when install dapp %v", err))
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

func (cmd *installCmd) helpMessage() {
	fmt.Printf(`
Usage:
	dapp-installer install [flags]

Flags:
	--config	Configuration file
	--help      Display help information
`)
}
