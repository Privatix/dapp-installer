package product

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
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
	var commandStr string
	var file string
	var arguments []string
	if err := json.Unmarshal([]byte("\""+v.Command+"\""), &commandStr); err != nil {
		return fmt.Errorf("could not parse command: %v", err)
	}
	// HACK(furkhat): mid stage migration to use <ROLE>, <OLD_PRODDIR> and <PRODDIR>.
	if strings.Contains(v.Command, "<OLD_PRODDIR>") || strings.Contains(v.Command, "<PRODDIR>") {
		s := strings.ReplaceAll(commandStr, "<OLD_PRODDIR>", oldProdDir)
		s = strings.ReplaceAll(s, "<PRODDIR>", prodDir)
		s = strings.ReplaceAll(s, "<ROLE>", role)
		s = strings.ReplaceAll(s, "'", "\"")
		arguments = strings.Split(s, " ")
		file = arguments[0]
	} else {
		arguments = strings.Split(v.Command, " ")
		for i, s := range arguments {
			arguments[i] = strings.Replace(s, "..", prodDir, -1)
		}
		file = filepath.Join(prodDir, arguments[0])
		if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			if _, err := os.Stat(file); err != nil {
				file = arguments[0]
			}
		}
	}

	var err error
	if v.Admin && runtime.GOOS == "darwin" {
		command := fmt.Sprintf("%s %s", file, strings.Join(arguments[1:], " "))
		err = util.ExecuteCommandOnDarwinAsAdmin(command)
	} else {
		err = util.ExecuteCommand(file, arguments[1:]...)
	}

	if err != nil {
		return fmt.Errorf("failed to execute %s: %v", strings.Join(arguments, " "), err)
	}
	return nil
}
