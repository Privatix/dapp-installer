package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/Privatix/dapp-installer/util"
	"github.com/Privatix/dappctrl/util/log"
)

func getUpdateCmd() *Command {
	return &Command{
		Name:    "update",
		execute: update,
	}
}

func update(conf *config, logger log.Logger) {
	h := flag.Bool("help", false, "Display dapp-installer help")

	flag.CommandLine.Parse(os.Args[2:])

	if *h {
		updateHelp()
		return
	}

	logger.Info("start update process")
	defer logger.Info("finish update process")

	fmt.Println("I will be running upgrading process")

	if !util.CheckSystemPrerequisites(logger) {
		logger.Warn("upgrading process was interrupted")
		return
	}
	logger.Info("check the system prerequisites was successful")
}

func updateHelp() {
	fmt.Print(`
Usage:
	dapp-installer update [flags]

Flags:
	--help      Display help information`)
}
