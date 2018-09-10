package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/Privatix/dapp-installer/data"

	"github.com/Privatix/dapp-installer/util"
	"github.com/Privatix/dappctrl/util/log"
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
  
Use "dapp-installer [command] --help" for more information about a command.`

// Command has a commands parameters
type Command struct {
	// Name of command
	Name string
	// execute has a pointer to execute func
	execute func(conf *config, logger log.Logger)
}

type config struct {
	DBEngine *util.DBEngine
	Registry []util.Registry
	DappCtrl *dappCtrlConfig
}

type dappCtrlConfig struct {
	File string
}

func newConfig() *config {
	return &config{
		DBEngine: &util.DBEngine{
			Download:    "https://get.enterprisedb.com/postgresql/postgresql-10.5-1-windows-x64.exe",
			ServiceName: "postrges",
			DB: &data.DB{
				DBName:   "dappctrl",
				User:     "postgres",
				Password: "postgres",
				Port:     "5432",
			},
		},
	}
}

// Execute has a command execute
func Execute(logger log.Logger, printVersion func(), args []string) {
	if len(args) == 0 {
		fmt.Print(mainHelpMsg)
		return
	}

	commands := map[string]*(Command){
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
	cmd.execute(conf, logger.Add("command", cmd.Name))
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
