package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/privatix/dappctrl/util/log"
)

type removeCmd struct {
	name          string
	rollbackFuncs []func(conf *config, logger log.Logger)
}

func getRemoveCmd() *removeCmd {
	return &removeCmd{name: "remove"}
}

func removeProcessedFlags(conf *config, logger log.Logger) bool {
	h := flag.Bool("help", false, "Display dapp-installer help")
	p := flag.String("workdir", ".", "Dapp install directory")

	flag.CommandLine.Parse(os.Args[2:])

	if *h {
		removeHelp()
		return true
	}

	conf.Dapp.SetInstallPath(*p)
	return false
}

func (cmd *removeCmd) execute(conf *config, log log.Logger) error {
	logger := log.Add("command", cmd.name)
	if removeProcessedFlags(conf, logger) {
		return nil
	}

	logger.Info("start process")
	defer logger.Info("finish process")

	existDapp, ok := existingDapp(conf.Dapp.InstallPath, logger)
	if !ok {
		msg := fmt.Sprintf("dapp is not found at %s", conf.Dapp.InstallPath)
		fmt.Println(msg)
		logger.Warn(msg)
		return nil
	}

	conf.Dapp = existDapp
	return remove(conf, logger)
}

func remove(conf *config, logger log.Logger) error {
	dapp := conf.Dapp

	if err := dapp.Remove(logger); err != nil {
		logger.Warn(fmt.Sprintf("failed to remove dapp: %v", err))
		return err
	}
	return nil
}

func (cmd *removeCmd) rollback(conf *config, logger log.Logger) {
	for i := len(cmd.rollbackFuncs) - 1; i >= 0; i-- {
		cmd.rollbackFuncs[i](conf, logger)
	}
}

func (cmd *removeCmd) addRollbackFuncs(f func(c *config, l log.Logger)) {
	cmd.rollbackFuncs = append(cmd.rollbackFuncs, f)
}

func removeHelp() {
	fmt.Print(`
Usage:
	dapp-installer remove [flags]

Flags:
	--help		Display help information
	--workdir	Dapp install directory
`)
}
