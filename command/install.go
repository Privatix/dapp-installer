package command

import (
	"flag"
	"fmt"
	"os"
	"strconv"

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

func installProcessedFlags(conf *config, log log.Logger) bool {
	h := flag.Bool("help", false, "Display dapp-installer help")
	configFile := flag.String("config", "", "Configuration file")

	flag.CommandLine.Parse(os.Args[2:])

	if *h {
		installHelp()
		return true
	}

	if *configFile != "" {
		if err := dapputil.ReadJSONFile(*configFile, &conf); err != nil {
			log.Error(fmt.Sprintf("failed to read config file - %s", err))
			return true
		}
	}
	return false
}

func install(conf *config, log log.Logger) {
	if installProcessedFlags(conf, log) {
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

	p, exist := util.DBEngineExists(log)
	if !exist {
		log.Warn("db engine is not exists")
		if err := util.InstallDBEngine(conf.DBEngine, log); err != nil {
			log.Warn(
				fmt.Sprintf("ocurred error while installing dbengine %v", err))
			return
		}
		p, _ = util.DBEngineExists(log)
	}

	conf.DBEngine.DB.Port = strconv.Itoa(p)

	// check to access db engine service
	connStr := util.GetConnectionString("postgres", conf.DBEngine.DB.User,
		conf.DBEngine.DB.Password, conf.DBEngine.DB.Port)
	if err := data.Ping(connStr); err != nil {
		log.Warn(fmt.Sprintf(
			"ocurred error when check to access dbengine service %v", err))
		return
	}

	dappConnStr := util.GetConnectionString(conf.DBEngine.DB.DBName,
		conf.DBEngine.DB.User, conf.DBEngine.DB.Password,
		conf.DBEngine.DB.Port)
	// check to access dapp database
	if err := data.Ping(dappConnStr); err != nil {
		// create dapp database
		err := createDatabase(conf.DBEngine.DB.DBName, connStr, dappConnStr)
		if err != nil {
			log.Warn(fmt.Sprintf(
				"ocurred error when create database %v", err))
			return
		}
		log.Info("database successfully created")
	}
	log.Info("checking the dbengine exists was successful")
}

func createDatabase(dbname, conn1, conn2 string) error {
	err := data.CreateDatabase(dbname, conn1)
	if err != nil {
		return err
	}
	return data.ConfigurateDatabase(conn2)
}

func installHelp() {
	fmt.Print(`
Usage:
	dapp-installer install [flags]

Flags:
	--help      Display help information`)
}
