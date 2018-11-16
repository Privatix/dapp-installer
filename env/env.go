package env

import (
	"encoding/json"
	"os"

	dapputil "github.com/privatix/dappctrl/util"
)

// Variables has a store environment variables.
type Variables struct {
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

// NewVariables creates a default Varibales configuration.
func NewVariables() *Variables {
	return &Variables{
		Schema: "1.0",
		Dapp:   &dapp{},
		DB:     &db{},
		Tor:    &tor{},
	}
}

// Write saves the variables to json file.
func (v *Variables) Write(path string) error {
	write, err := os.Create(path)
	if err != nil {
		return err
	}
	defer write.Close()

	return json.NewEncoder(write).Encode(v)
}

// Read reads the variables from json file.
func (v *Variables) Read(path string) error {
	return dapputil.ReadJSONFile(path, &v)
}
