package tor

import (
	"os"
	"path/filepath"
	"text/template"

	"github.com/privatix/dapp-installer/statik"
	"github.com/privatix/dapp-installer/util"
)

// Tor has a tor configuration.
type Tor struct {
	HiddenServiceDir string
	RootPath         string
	SocksPort        int
	VirtPort         int
	TargetPort       int
}

// NewTor creates a default Tor configuration.
func NewTor() *Tor {
	return &Tor{
		HiddenServiceDir: "tor/hidden_service",
		SocksPort:        9999,
		VirtPort:         80,
		TargetPort:       80,
	}
}

func (t *Tor) configurate() error {
	//todo create private key and hostname

	return t.createSettings()
}

func (t *Tor) createSettings() error {
	file, err := os.Create(filepath.Join(t.RootPath, "tor/settings/torrc"))
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := statik.ReadFile("/templates/torrc.tpl")
	if err != nil {
		return err
	}

	templ, err := template.New("torTemplate").Parse(string(data))
	if err != nil {
		return err
	}

	// todo set dynamic port

	return templ.Execute(file, &t)
}

// Install installs tor process.
func (t *Tor) Install() error {
	if err := t.configurate(); err != nil {
		return err
	}

	return installService(t.serviceName(), t.RootPath)
}

// Start starts tor process.
func (t *Tor) Start() error {
	return startService(t.serviceName())
}

// Stop stops tor process.
func (t *Tor) Stop() error {
	if util.IsServiceStopped(t.serviceName()) {
		return nil
	}
	return stopService(t.serviceName())
}

// Remove removes tor process.
func (t *Tor) Remove() error {
	return removeService(t.serviceName())
}
