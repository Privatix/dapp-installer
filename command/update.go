package command

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/privatix/dapp-installer/data"
	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dappctrl/util/log"
)

type updateCmd struct {
	name          string
	rollbackFuncs []func(conf *config, logger log.Logger)
}

func getUpdateCmd() *updateCmd {
	return &updateCmd{name: "update"}
}

func updateHelpMessage() {
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
	if commandProcessedFlags(updateHelpMessage, conf, logger) {
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
	oldDapp, ok := existingDapp(conf.Dapp.UserRole, logger)

	if !ok {
		fmt.Println("dapp is not installed.")
		fmt.Println("to install run 'install' command")
		logger.Warn("dapp is not installed")
		return nil
	}

	b, dir := filepath.Split(oldDapp.InstallPath)
	conf.Dapp.InstallPath = filepath.Join(b, dir+"_new")

	newDapp, err := initDapp(conf)
	if err != nil {
		return err
	}

	version := util.ParseVersion(newDapp.Version)
	if util.ParseVersion(oldDapp.Version) >= version {
		fmt.Printf("dapp current version: %s, update is not required\n",
			oldDapp.Version)
		logger.Warn("already updated")
		return nil
	}

	configFile := filepath.Join(oldDapp.InstallPath,
		oldDapp.Controller.Configuration)
	if newDapp.DBEngine.DB, err = dbParamsFromConfig(configFile); err != nil {
		logger.Warn(fmt.Sprintf("failed to read config: %v", err))
		return err
	}

	// Update dapp core.
	if err := newDapp.Update(oldDapp, logger); err != nil {

		oldDapp.DBEngine.Start(oldDapp.InstallPath)
		oldDapp.Controller.Service.Start()

		logger.Warn(fmt.Sprintf("failed to update dapp: %v", err))
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
