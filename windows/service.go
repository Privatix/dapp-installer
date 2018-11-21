// +build windows

//go:generate goversioninfo -64 -o ../resource_windows_amd64.syso

package windows

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
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

	fileName := filepath.Join(path, s.ID+".exe")
	err = ioutil.WriteFile(fileName, bytes, 0644)
	if err != nil {
		return err
	}
	// create winservice wrapper config file
	s.Command = filepath.Join(path, s.Command)
	for i, arg := range s.Args {
		if strings.HasSuffix(arg, ".json") {
			s.Args[i] = filepath.Join(path, arg)
		}
	}
	bytes, err = json.Marshal(s)
	if err != nil {
		return err
	}
	fileName = filepath.Join(path, s.ID+".config.json")
	return ioutil.WriteFile(fileName, bytes, 0644)
}
