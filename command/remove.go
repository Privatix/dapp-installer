package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/Privatix/dappctrl/util/log"
)

func getRemoveCmd() *Command {
	return &Command{
		Name:    "remove",
		execute: remove,
	}
}

func remove(conf *config, logger log.Logger) {
	h := flag.Bool("help", false, "Display dapp-installer help")

	flag.CommandLine.Parse(os.Args[2:])

	if *h {
		removeHelp()
		return
	}

	logger.Info("start remove process")
	fmt.Println("I will be running uninstallation process")
	logger.Info("finish remove process")
}

func removeHelp() {
	fmt.Print(`
Usage:
	dapp-installer remove [flags]

Flags:
	--help      Display help information`)
}
