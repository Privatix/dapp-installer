package command

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/privatix/dapp-installer/dapp"
	"github.com/privatix/dapp-installer/util"
	dapputil "github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

const mainHelpMsg = `
Usage:
  dapp-installer [command] [flags]

Available Commands:
  install     Install dapp core
  update      Update dapp core
  remove      Remove dapp core

Flags:
  --help      Display help information
  --version   Display the current version of this CLI

Use "dapp-installer [command] --help" for more information about a command.
`

type command interface {
	execute(conf *config, logger log.Logger) error
	rollback(conf *config, logger log.Logger)
}

type config struct {
	// Dapp contains a dapp installation parameters.
	Dapp *dapp.Dapp
}

func newConfig() *config {
	return &config{
		Dapp: dapp.NewConfig(),
	}
}

// Execute has a command execute
func Execute(logger log.Logger, printVersion func(), args []string) {
	if len(args) == 0 {
		fmt.Print(mainHelpMsg)
		return
	}

	commands := map[string]command{
		"install": getInstallCmd(),
		"update":  getUpdateCmd(),
		"remove":  getRemoveCmd(),
	}

	cmd, exist := commands[strings.ToLower(args[0])]
	if !exist {
		if !processedFlags(printVersion) {
			fmt.Println("unknown command or flags:", args[0])
			fmt.Print(mainHelpMsg)
		}
		return
	}

	conf := newConfig()
	if err := cmd.execute(conf, logger); err != nil {
		cmd.rollback(conf, logger)
		fmt.Println("installation was canceled after an exception occurred.")
		logger.Info("installation was canceled")
		return
	}
	fmt.Println("command successfully finished")
}

func processedFlags(printVersion func()) bool {
	v := flag.Bool("version", false, "Prints current dapp-installer version")
	h := flag.Bool("help", false, "Display dapp-installer help")

	flag.Parse()

	if *v {
		printVersion()
		return true
	}
	if *h {
		fmt.Printf(mainHelpMsg)
		return true
	}
	return false
}

func initDapp(conf *config) (*dapp.Dapp, error) {
	d := conf.Dapp
	downloadPath := conf.Dapp.Download()

	ch := make(chan bool)
	defer close(ch)
	go util.InteractiveWorker("Extracting dapp", ch)

	if _, err := os.Stat(d.InstallPath); os.IsNotExist(err) {
		os.MkdirAll(d.InstallPath, util.FullPermission)
	}

	_, err := util.Unzip(downloadPath, d.InstallPath)
	if err != nil {
		ch <- true
		return nil, err
	}

	fileName := filepath.Join(d.InstallPath, d.Controller.EntryPoint)
	d.Version = util.DappCtrlVersion(fileName)

	ch <- true
	fmt.Printf("\r%s\n", "Dapp was successfully extracted")

	return d, nil
}

func uninstallDapp(conf *config, logger log.Logger) {
	err := conf.Dapp.Remove(logger)
	if err != nil {
		logger.Warn(fmt.Sprintf("failed to remove dapp: %v", err))
		return
	}

	logger.Info("dappctrl was successfully removed")
}

func existingDapp(role string, logger log.Logger) (*dapp.Dapp, bool) {
	return dapp.Exists(role, logger)
}

func commandProcessedFlags(help func(), conf *config,
	logger log.Logger) bool {
	h := flag.Bool("help", false, "Display dapp-installer help")
	configFile := flag.String("config", "", "Configuration file")

	flag.CommandLine.Parse(os.Args[2:])

	if *h {
		help()
		return true
	}

	if *configFile == "" {
		logger.Warn("config parameter is empty")
		fmt.Println("config parameter is empty")
		return true
	}
	if err := dapputil.ReadJSONFile(*configFile, &conf); err != nil {
		logger.Error(fmt.Sprintf("failed to read config: %s", err))
		return true
	}
	return false
}
