package product

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
)

const (
	productsDir      = "product"
	agentConfigPath  = "config/install.agent.config.json"
	clientConfigPath = "config/install.client.config.json"
)

// StopAll stops all products.
func StopAll(ctx context.Context, logger log.Logger, installRoot, role string) error {
	return runForAllProducts(ctx, logger, "", installRoot, role, "stop")
}

// StartAll starts all products.
func StartAll(ctx context.Context, logger log.Logger, installRoot, role string) error {
	return runForAllProducts(ctx, logger, "", installRoot, role, "start")
}

// UpdateAll updates all products.
func UpdateAll(ctx context.Context, logger log.Logger, oldInstallRoot, installRoot, role string) error {
	return runForAllProducts(ctx, logger, oldInstallRoot, installRoot, role, "update")
}

func runForAllProducts(ctx context.Context, logger log.Logger, oldInstallRoot, installRoot, role, action string) error {
	logger.Info("list all products")
	products, err := ioutil.ReadDir(filepath.Join(installRoot, productsDir))
	if err != nil {
		return fmt.Errorf("could not list installed products: %v", err)
	}

	logger.Info("read install configs")
	// Get all products install.<role>.config.json
	allConfigs := make([]installConfig, 0)
	configsLocations := make([]string, 0)
	for _, p := range products {
		if !p.IsDir() {
			continue
		}
		configPath := agentConfigPath
		if role == data.RoleClient {
			configPath = clientConfigPath
		}
		var c installConfig
		location := filepath.Join(installRoot, productsDir, p.Name(), configPath)
		if err := util.ReadJSON(location, &c); err != nil {
			return fmt.Errorf("could not read products config: %v", err)
		}
		configsLocations = append(configsLocations, location)
		allConfigs = append(allConfigs, c)
	}

	logger.Info("executing " + action + " products")
	// Execute all stop commands from all configs.
	for i, conf := range allConfigs {
		confLocation := configsLocations[i]
		productDir, _ := filepath.Abs(filepath.Join(filepath.Dir(confLocation), ".."))
		oldProductDir := filepath.Join(oldInstallRoot, "product", filepath.Base(productDir))
		commands := conf.Start
		if action == "start" {
		} else if action == "stop" {
			commands = conf.Stop
		} else if action == "update" {
			commands = conf.Update
		}
		for _, v := range commands {
			if err := executeCommand(oldProductDir, productDir, role, v); err != nil {
				return err
			}
		}
	}
	return nil
}

func executeCommand(oldProdDir, prodDir, role string, v command) error {
	commandStr := v.Command
	// HACK(furkhat): mid stage migration to use <ROLE>, <OLD_PRODDIR> and <PRODDIR>.
	if strings.Contains(v.Command, "<OLD_PRODDIR>") || strings.Contains(v.Command, "<PRODDIR>") {
		commandStr = strings.ReplaceAll(commandStr, "<OLD_PRODDIR>", oldProdDir)
		commandStr = strings.ReplaceAll(commandStr, "<PRODDIR>", prodDir)
		commandStr = strings.ReplaceAll(commandStr, "<ROLE>", role)
	}

	var err error
	if runtime.GOOS == "windows" {
		err = util.ExecuteCommand("powershell", "-ExecutionPolicy", "Bypass",
			"-Command", commandStr)
	} else {
		if v.Admin && runtime.GOOS == "darwin" {
			err = util.ExecuteCommandOnDarwinAsAdmin(commandStr)
		} else {
			err = util.ExecuteCommand("sh", "-c", commandStr)
		}
	}
	if err != nil {
		return fmt.Errorf("failed to execute %s: %v", commandStr, err)
	}
	return nil
}
