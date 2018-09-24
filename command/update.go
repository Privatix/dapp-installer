package command

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/privatix/dapp-installer/data"
	"github.com/privatix/dapp-installer/util"
	dapputil "github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

type updateCmd struct {
	name          string
	rollbackFuncs []func(conf *config, logger log.Logger)
}

func getUpdateCmd() *updateCmd {
	return &updateCmd{name: "update"}
}

func updateProcessedFlags(cmd *updateCmd, conf *config, logger log.Logger) bool {
	h := flag.Bool("help", false, "Display dapp-installer help")
	configFile := flag.String("config", "", "Configuration file")

	flag.CommandLine.Parse(os.Args[2:])

	if *h {
		cmd.helpMessage()
		return true
	}

	if *configFile == "" {
		logger.Warn("config parameter is empty")
		fmt.Println("config parameter is empty")
		return true
	}
	if err := dapputil.ReadJSONFile(*configFile, &conf); err != nil {
		logger.Error(fmt.Sprintf("failed to read config file - %s", err))
		return true
	}
	return false
}

func (cmd *updateCmd) helpMessage() {
	fmt.Printf(`
Usage:
	dapp-installer update [flags]

Flags:
	--config	Configuration file
	--help      Display help information
`)
}

func (cmd *updateCmd) execute(conf *config, log log.Logger) error {
	logger := log.Add("command", cmd.name)
	if updateProcessedFlags(cmd, conf, logger) {
		return nil
	}

	logger.Info("start process")
	defer logger.Info("finish process")

	volume := filepath.VolumeName(conf.Dapp.InstallPath)
	conf.Dapp.TempPath = util.TempPath(volume)
	defer os.RemoveAll(conf.Dapp.TempPath)

	return updateDapp(cmd, conf, logger)
}

func updateDapp(cmd *updateCmd, conf *config, logger log.Logger) error {
	existDapp, ok := existingDapp(conf.Dapp.UserRole, logger)

	if !ok {
		fmt.Println("dapp is not exist.")
		fmt.Println("for install you should run with command 'install'")
		logger.Warn("dapp is not exists")
		return nil
	}

	b, dir := filepath.Split(existDapp.InstallPath)
	conf.Dapp.InstallPath = filepath.Join(b, dir+"_new")

	newDapp, err := initDapp(conf)
	if err != nil {
		return err
	}

	version := util.ParseVersion(newDapp.Version)
	if util.ParseVersion(existDapp.Version) >= version {
		fmt.Printf("you don't need to update. your dapp version: %v\n",
			existDapp.Version)
		logger.Warn("don't need to update")
		return nil
	}

	configFile := filepath.Join(existDapp.InstallPath,
		existDapp.Controller.Configuration)
	if newDapp.DBEngine.DB, err = dbParamsFromConfig(configFile); err != nil {
		logger.Warn(fmt.Sprintf("ocurred error when read config %v", err))
		return err
	}

	// update dapp core
	if err := newDapp.Update(existDapp, logger); err != nil {

		existDapp.DBEngine.Start(existDapp.InstallPath)
		existDapp.Controller.Service.Start()

		logger.Warn(fmt.Sprintf("ocurred error when update dapp %v", err))
		return err
	}
	cmd.addRollbackFuncs(uninstallDapp)

	return nil
}

func (cmd *updateCmd) addRollbackFuncs(f func(c *config, l log.Logger)) {
	cmd.rollbackFuncs = append(cmd.rollbackFuncs, f)
}

func (cmd *updateCmd) rollback(conf *config, logger log.Logger) {
	for i := len(cmd.rollbackFuncs) - 1; i >= 0; i-- {
		cmd.rollbackFuncs[i](conf, logger)
	}
}

func dbParamsFromConfig(configFile string) (*data.DB, error) {
	read, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer read.Close()

	jsonMap := make(map[string]interface{})

	json.NewDecoder(read).Decode(&jsonMap)

	db, ok := jsonMap["DB"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("DB params not found")
	}
	conn, ok := db["Conn"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Conn params not found")
	}

	res := data.NewConfig()

	if dbname, ok := conn["dbname"]; ok {
		res.DBName = dbname.(string)
	}

	if user, ok := conn["user"]; ok {
		res.User = user.(string)
	}

	if pwd, ok := conn["password"]; ok {
		res.Password = pwd.(string)
	}

	if port, ok := conn["port"]; ok {
		res.Port = port.(string)
	}

	return res, nil
}
