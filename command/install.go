package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/Privatix/dapp-installer/util"
	"github.com/Privatix/dappctrl/util/log"
)

func getInstallCmd() *Command {
	return &Command{
		Name:    "install",
		execute: install,
	}
}

func install(log log.Logger) {
	h := flag.Bool("help", false, "Display dapp-installer help")

	flag.CommandLine.Parse(os.Args[2:])

	if *h {
		installHelp()
		return
	}

	log.Info("start install process")
	defer log.Info("finish install process")

	fmt.Println("I will be running installation process")

	if !util.CheckSystemPrerequisites(log) {
		log.Warn("installation process was interrupted")
		return
	}
	log.Info("check the system prerequisites was successful")
}

func installHelp() {
	fmt.Print(`
Usage:
	dapp-installer install [flags]

Flags:
	--help      Display help information`)
}
