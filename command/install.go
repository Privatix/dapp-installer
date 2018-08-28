package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/Privatix/dapp-installer/data"
	"github.com/Privatix/dapp-installer/util"
	dapputil "github.com/Privatix/dappctrl/util"
	"github.com/Privatix/dappctrl/util/log"
)

func getInstallCmd() *Command {
	return &Command{
		Name:    "install",
		execute: install,
	}
}

func install(conf *config, log log.Logger) {
	h := flag.Bool("help", false, "Display dapp-installer help")
	configFile := flag.String("config", "", "Configuration file")

	flag.CommandLine.Parse(os.Args[2:])

	if *h {
		installHelp()
		return
	}

	if *configFile != "" {
		if err := dapputil.ReadJSONFile(*configFile, &conf); err != nil {
			log.Error(fmt.Sprintf("failed to read config file - %s", err))
			return
		}
	}

	log.Info("start install process")
	defer log.Info("finish install process")

	fmt.Println("I will be running installation process")

	if !util.CheckSystemPrerequisites(log) {
		log.Warn("installation process was interrupted")
		return
	}
	log.Info("check the system prerequisites was successful")

	if !util.DBEngineExists(log) {
		log.Warn("db engine is not exists")
		if err := util.InstallDBEngine(conf.DBEngine, log); err != nil {
			log.Warn(
				fmt.Sprintf("ocurred error while installing dbengine %v", err))
			return
		}
		// check to access db engine service
		connStr := util.GetConnectionString("postgres", conf.DBEngine.DB.User,
			conf.DBEngine.DB.Password, conf.DBEngine.DB.Port)

		if err := data.Ping(connStr); err != nil {
			log.Warn(fmt.Sprintf(
				"ocurred error when check to access dbengine service %v", err))
			return
		}
	}
	log.Info("checking the dbengine exists was successful")
}

func installHelp() {
	fmt.Print(`
Usage:
	dapp-installer install [flags]

Flags:
	--help      Display help information`)
}
