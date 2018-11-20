package product

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	productDir     = "product"
	envFile        = ".env"
	configFile     = "install.config.json"
	productImport  = "PRODUCT_IMPORT"
	productInstall = "PRODUCT_INSTALL"
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

func writeVariable(envFile, key string, value string) error {

	env, err := godotenv.Read(envFile)
	if err != nil {
		return err
	}

	env[key] = value

	if err := godotenv.Write(env, envFile); err != nil {
		return fmt.Errorf("failed to write env file: %v", err)
	}

	return nil
}

func readVariable(envFile, key string) (string, bool) {
	env, err := godotenv.Read(envFile)
	if err != nil {
		return "", false
	}

	value, ok := env[key]
	return value, ok
}

func getParameters(path string) (string, bool, bool) {
	var imported, installed bool
	envPath, _ := findFile(path, envFile)

	if len(envPath) == 0 {
		envPath = filepath.Join(path, "config/"+envFile)
	}

	v, ok := readVariable(envPath, productImport)
	if ok {
		imported, _ = strconv.ParseBool(v)
	}

	v, ok = readVariable(envPath, productInstall)
	if ok {
		installed, _ = strconv.ParseBool(v)
	}
	return envPath, imported, installed
}
