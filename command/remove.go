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

func (cmd *removeCmd) execute(conf *config, logger log.Logger) error {
	h := flag.Bool("help", false, "Display dapp-installer help")

	flag.CommandLine.Parse(os.Args[2:])

	if *h {
		removeHelp()
		return nil
	}

	logger.Info("start remove process")
	fmt.Println("I will be running uninstallation process")
	logger.Info("finish remove process")
	return nil
}

func (cmd *removeCmd) rollback(conf *config, logger log.Logger) {
	for i := len(cmd.rollbackFuncs) - 1; i >= 0; i-- {
		cmd.rollbackFuncs[i](conf, logger)
	}
}

func removeHelp() {
	fmt.Print(`
Usage:
	dapp-installer remove [flags]

Flags:
	--help      Display help information`)
}
