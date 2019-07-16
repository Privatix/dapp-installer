// +build windows

package dapp

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/privatix/dapp-installer/dbengine"
	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dapp-installer/windows"
)

type service struct {
	windows.Service
}

// SetUID needed only on darwin, does nothing.
func (*service) SetUID(_ string) {}

func (s *service) Start() error {
	err := util.StartService(s.Name)
	if err != nil {
		err = s.Service.Start()
	}
	return err
}

func (s *service) Stop() error {
	err := util.ExecuteCommand("sc", "stop", s.Name)
	if err != nil {
		err = s.Service.Stop()
	}
	return err
}

func newService() *service {
	return &service{
		windows.Service{
			Command:   "dappctrl",
			Args:      []string{"-config", "dappctrl.config.json"},
			AutoStart: true,
		},
	}
}

// ControllerHash unique name for installing path.
func (d *Dapp) ControllerHash() string {
	return fmt.Sprintf("Privatix Controller %s", util.Hash(d.Path))
}

// Configure configurates installed dapp.
func (d *Dapp) Configure() error {
	ctrl := d.Controller
	ctrlPath := filepath.Join(d.Path, filepath.Dir(ctrl.EntryPoint))

	hash := d.ControllerHash()
	ctrl.Service.ID = hash
	ctrl.Service.AutoStart = d.Role == "agent"
	ctrl.Service.Name = hash
	ctrl.Service.Description = fmt.Sprintf("dapp controller %s", hash)
	ctrl.Service.GUID = filepath.Join(ctrlPath, ctrl.Service.ID)
	if err := ctrl.Service.CreateWrapper(ctrlPath); err != nil {
		return fmt.Errorf("failed to create service wrapper: %v", err)
	}

	if err := ctrl.Service.Install(); err != nil {
		return fmt.Errorf("failed to install service: %v", err)
	}

	if err := d.modifyDappConfig(); err != nil {
		return err
	}

	if err := ctrl.Service.Start(); err != nil {
		// retry attempt to start service
		if err := ctrl.Service.Start(); err != nil {
			return fmt.Errorf("failed to start service: %v", err)
		}
	}

	return configurateService(d)
}

func configurateService(d *Dapp) error {
	fail := func(name string) error {
		return util.ExecuteCommand("sc", "failure", name, "reset=", "0",
			"actions=", "restart/1000/restart/2000/restart/5000")
	}

	descr := func(name, descr string) error {
		role := d.Role
		hash := util.Hash(d.Path)
		return util.ExecuteCommand("sc", "description", name,
			fmt.Sprintf("Privatix %s %s %s", role, descr, hash))
	}

	services := []string{
		d.Controller.Service.ID,
		dbengine.Hash(d.Path),
	}
	suffixes := []string{"controller", "database"}

	for i, v := range services {
		if err := fail(v); err != nil {
			return err
		}

		if err := descr(v, suffixes[i]); err != nil {
			return err
		}
	}

	return util.ExecuteCommand("sc", "config", services[0],
		"depend=", services[1])
}

func copyServiceWrapper(d, s *Dapp) {
	hash := s.ControllerHash()
	scvExe := "dappctrl/" + hash + ".exe"
	scvConfig := "dappctrl/" + hash + ".config.json"

	util.CopyFile(filepath.Join(s.Path, scvExe),
		filepath.Join(d.Path, scvExe))
	util.CopyFile(filepath.Join(s.Path, scvConfig),
		filepath.Join(d.Path, scvConfig))
}

func clearGuiStorage() error {
	output, err := util.ExecuteCommandOutput("wmic", "useraccount",
		"where", "Status='OK'", "get", "Name", "/value")

	if err != nil {
		return err
	}

	users := strings.Split(strings.TrimSpace(output), "Name=")

	for _, username := range users {
		username = strings.TrimSpace(username)
		if len(username) == 0 {
			continue
		}
		u, err := user.Lookup(username)
		if err != nil {
			return err
		}
		path := filepath.Join(u.HomeDir, "AppData", "Roaming", "dappctrlgui")
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}

	return nil
}
