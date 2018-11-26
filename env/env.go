package env

import (
	"encoding/json"
	"os"

	dapputil "github.com/privatix/dappctrl/util"
)

// Config has a store environment variables.
type Config struct {
	Schema  string
	Role    string
	WorkDir string
	Version string
	Dapp    *dapp
	DB      *db
	Tor     *tor
}

type dapp struct {
	Controller string
	Gui        string
	Service    string
}

type db struct {
	Service string
}

type tor struct {
	Service string
}

// NewConfig creates a default Configs configuration.
func NewConfig() *Config {
	return &Config{
		Schema: "1.0",
		Dapp:   &dapp{},
		DB:     &db{},
		Tor:    &tor{},
	}
}

// Write saves the configs to json file.
func (c *Config) Write(path string) error {
	write, err := os.Create(path)
	if err != nil {
		return err
	}
	defer write.Close()

	return json.NewEncoder(write).Encode(c)
}

// Read reads the configs from json file.
func (c *Config) Read(path string) error {
	return dapputil.ReadJSONFile(path, &c)
}
