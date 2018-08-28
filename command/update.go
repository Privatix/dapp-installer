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

func update(log log.Logger) {
	h := flag.Bool("help", false, "Display dapp-installer help")

	flag.CommandLine.Parse(os.Args[2:])

	if *h {
		updateHelp()
		return
	}

	log.Info("start update process")
	defer log.Info("finish update process")

	fmt.Println("I will be running upgrading process")

	if !util.CheckSystemPrerequisites(log) {
		log.Warn("upgrading process was interrupted")
		return
	}
	log.Info("check the system prerequisites was successful")
}

func updateHelp() {
	fmt.Print(`
Usage:
	dapp-installer update [flags]

Flags:
	--help      Display help information`)
}
