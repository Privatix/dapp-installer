package command

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/privatix/dapp-installer/data"
	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dappctrl/util/log"
)

type removeCmd struct {
	name          string
	rollbackFuncs []func(conf *config, logger log.Logger)
}

func getRemoveCmd() *removeCmd {
	return &removeCmd{name: "remove"}
}

func removeProcessedFlags(conf *config, logger log.Logger) bool {
	h := flag.Bool("help", false, "Display dapp-installer help")
	r := flag.String("role", "", "Dapps role")

	flag.CommandLine.Parse(os.Args[2:])

	if *h {
		removeHelp()
		return true
	}

	if r == nil || *r == "" {
		logger.Warn("role parameter is empty")
		fmt.Println("role parameter is empty")
		return true
	}

	conf.Dapp.UserRole = *r
	return false
}

func (cmd *removeCmd) execute(conf *config, log log.Logger) error {
	logger := log.Add("command", cmd.name)
	if removeProcessedFlags(conf, logger) {
		return nil
	}

	logger.Info("start process")
	defer logger.Info("finish process")

	existDapp, ok := existingDapp(conf.Dapp.UserRole, logger)
	if !ok {
		msg := fmt.Sprintf("dapp (%s) is not found", conf.Dapp.UserRole)
		fmt.Println(msg)
		logger.Warn(msg)
		return nil
	}

	conf.Dapp = existDapp
	return remove(conf, logger)
}

func remove(conf *config, logger log.Logger) error {
	dapp := conf.Dapp

	db, err := dbParamsFromConfig(dapp.InstallPath + dapp.Configuration)
	if err != nil {
		logger.Warn(fmt.Sprintf(
			"ocurred error when read config %v", err))
		return err
	}

	if data.DBExists(db, logger) {
		dapp.Service.Uninstall()

		if err := data.DropDatabase(db); err != nil {
			logger.Warn(fmt.Sprintf(
				"ocurred error when drop database %v", err))
			return err
		}
	}

	if err := dapp.Remove(); err != nil {
		logger.Warn(fmt.Sprintf(
			"ocurred error when remove dapp %v", err))
		return err
	}

	util.RemoveRegistryKey(conf.Registry, dapp.UserRole)

	return nil
}

func (cmd *removeCmd) rollback(conf *config, logger log.Logger) {
	for i := len(cmd.rollbackFuncs) - 1; i >= 0; i-- {
		cmd.rollbackFuncs[i](conf, logger)
	}
}

func (cmd *removeCmd) addRollbackFuncs(f func(c *config, l log.Logger)) {
	cmd.rollbackFuncs = append(cmd.rollbackFuncs, f)
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

	res := &data.DB{
		DBName:   conn["dbname"].(string),
		User:     conn["user"].(string),
		Password: conn["password"].(string),
		Port:     conn["port"].(string),
	}
	return res, nil
}

func removeHelp() {
	fmt.Print(`
Usage:
	dapp-installer remove [flags]

Flags:
	--help	Display help information
	--role	Dapp for selected role will be removed
`)
}
