package command

import (
	"flag"
	"fmt"
	"os"

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

func (cmd *updateCmd) rollback(conf *config, logger log.Logger) {
	for i := len(cmd.rollbackFuncs) - 1; i >= 0; i-- {
		cmd.rollbackFuncs[i](conf, logger)
	}
}

func (cmd *updateCmd) execute(conf *config, logger log.Logger) error {
	h := flag.Bool("help", false, "Display dapp-installer help")

	flag.CommandLine.Parse(os.Args[2:])

	if *h {
		updateHelp()
		return nil
	}

	logger.Info("start update process")
	defer logger.Info("finish update process")

	fmt.Println("I will be running upgrading process")

	if !util.CheckSystemPrerequisites(logger) {
		logger.Warn("upgrading process was interrupted")
		return nil
	}
	logger.Info("check the system prerequisites was successful")
	return nil
}

func updateHelp() {
	fmt.Print(`
Usage:
	dapp-installer update [flags]

Flags:
	--help      Display help information`)
}
