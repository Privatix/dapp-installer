package product

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const (
	productDir       = "product"
	envFile          = ".env.product.config.json"
	agentConfigFile  = "install.agent.config.json"
	clientConfigFile = "install.client.config.json"
	productImport    = "ProductImport"
	productInstall   = "ProductInstall"
)

func findFile(dir, file string) (string, error) {
	var filePath string
	err := filepath.Walk(dir,
		func(p string, f os.FileInfo, err error) error {
			if strings.EqualFold(f.Name(), file) {
				filePath = p
			}
			return nil
		})

	return filePath, err
}

func writeVariable(envFile, key string, value interface{}) error {
	jsonMap := make(map[string]interface{})

	if read, err := os.Open(envFile); err == nil {
		json.NewDecoder(read).Decode(&jsonMap)
		read.Close()
	}

	jsonMap[key] = value
	write, err := os.Create(envFile)
	if err != nil {
		return err
	}
	defer write.Close()

	return json.NewEncoder(write).Encode(jsonMap)
}

func readVariable(envFile, key string) (interface{}, bool) {
	read, err := os.Open(envFile)
	if err != nil {
		return nil, false
	}
	defer read.Close()
	jsonMap := make(map[string]interface{})
	json.NewDecoder(read).Decode(&jsonMap)

	value, ok := jsonMap[key]
	if !ok {
		return nil, false
	}
	return value, true
}

func getParameters(path string) (string, bool, bool) {
	var imported, installed bool
	envPath, _ := findFile(path, envFile)

	if len(envPath) == 0 {
		envPath = filepath.Join(path, "config", envFile)
	}

	v, ok := readVariable(envPath, productImport)
	if ok {
		imported, _ = v.(bool)
	}

	v, ok = readVariable(envPath, productInstall)
	if ok {
		installed, _ = v.(bool)
	}
	return envPath, imported, installed
}

func filterProducts(files []os.FileInfo, product string) []os.FileInfo {
	if len(product) == 0 {
		return files
	}

	for _, f := range files {
		if f.Name() == product {
			return []os.FileInfo{f}
		}
	}
	return nil
}
