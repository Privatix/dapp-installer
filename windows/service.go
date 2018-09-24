// +build windows

package windows

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/privatix/dapp-installer/statik"
	"github.com/privatix/dapp-installer/util"
)

// Service has a parameters for windows service.
type Service struct {
	ID          string
	GUID        string
	Name        string
	Description string
	Command     string
	Args        []string
	AutoStart   bool
}

// Install installs windows service.
func (s *Service) Install() error {
	return util.ExecuteCommand(s.GUID, []string{"install"})
}

// Uninstall uninstalls windows service.
func (s *Service) Uninstall() {
	s.Stop()
	s.Remove()
}

// Start starts windows service.
func (s *Service) Start() error {
	return util.ExecuteCommand(s.GUID, []string{"start"})
}

// Stop stops windows service.
func (s *Service) Stop() error {
	return util.ExecuteCommand(s.GUID, []string{"stop"})
}

// Remove removes windows service.
func (s *Service) Remove() error {
	return util.ExecuteCommand(s.GUID, []string{"remove"})
}

// CreateWrapper creates service wrapper.
func (s *Service) CreateWrapper(path string) error {
	// create winservice wrapper file
	bytes, err := statik.ReadFile("/wrapper/winsvc.exe")
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0644)
	}

	err = ioutil.WriteFile(path+s.ID+".exe", bytes, 0644)
	if err != nil {
		return err
	}
	// create winservice wrapper config file
	s.Command = path + s.Command
	for i, arg := range s.Args {
		if strings.HasSuffix(arg, ".json") {
			s.Args[i] = path + arg
		}
	}
	bytes, err = json.Marshal(s)
	if err != nil {
		fmt.Println("error:", err)
	}
	return ioutil.WriteFile(path+s.ID+".config.json", bytes, 0644)
}
