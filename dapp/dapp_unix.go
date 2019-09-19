// +build linux darwin

package dapp

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/privatix/dapp-installer/unix"
	"github.com/privatix/dapp-installer/util"
)

type service struct {
	unix.Daemon
}

func newService() *service {
	return &service{
		unix.Daemon{
			Command: "dappctrl",
			Args:    []string{"-config", "dappctrl.config.json"},
		},
	}
}

// ControllerHash unique name for installing path.
func (d *Dapp) ControllerHash() string {
	return fmt.Sprintf("ctrl_%s", util.Hash(d.Path))
}

// Configure configures installed dapp.
func (d *Dapp) Configure() error {
	if err := d.modifyDappConfig(); err != nil {
		return err
	}

	ctrl := d.Controller

	ctrl.Service.ID = d.ControllerHash()
	ctrl.Service.AutoStart = d.Role == "agent"

	ctrl.Service.Command = filepath.Join(d.Path, ctrl.EntryPoint)
	ctrl.Service.Args = []string{
		"-config", filepath.Join(d.Path, ctrl.Configuration)}
	ctrl.Service.UID = d.UID

	if err := ctrl.Service.Install(); err != nil {
		return fmt.Errorf("failed to install daemon: %v", err)
	}
	if err := ctrl.Service.Start(); err != nil {
		// retry attempt to start service
		if err := ctrl.Service.Start(); err != nil {
			return fmt.Errorf("failed to start daemon: %v", err)
		}
	}

	return nil
}

func copyServiceWrapper(d, s *Dapp) {
}

func clearGuiStorage() error {
	u, err := user.Current()
	if err != nil {
		return err
	}
	path := u.HomeDir
	if runtime.GOOS == "darwin" {
		path = filepath.Join(path, "Library", "Application Support",
			"dappctrlgui")
	} else {
		path = filepath.Join(os.Getenv("HOME"), ".config", "dappctrlgui")
	}

	return os.RemoveAll(path)
}
