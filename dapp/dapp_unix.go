// +build linux darwin

package dapp

import (
	"fmt"
	"os"
	"path/filepath"

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

// Configurate configurates installed dapp.
func (d *Dapp) Configurate() error {
	if err := d.modifyDappConfig(); err != nil {
		return err
	}

	ctrl := d.Controller

	ctrl.Service.ID = d.controllerHash()

	ctrl.Service.Command = filepath.Join(d.Path, ctrl.EntryPoint)
	ctrl.Service.Args = []string{
		"-config", filepath.Join(d.Path, ctrl.Configuration)}

	if err := ctrl.Service.Install(); err != nil {
		return fmt.Errorf("failed to install daemon: %v", err)
	}
	if err := ctrl.Service.Start(); err != nil {
		// retry attempt to start service
		if err := ctrl.Service.Start(); err != nil {
			return fmt.Errorf("failed to start daemon: %v", err)
		}
	}

	_, installer := filepath.Split(os.Args[0])
	return util.CopyFile(os.Args[0], filepath.Join(d.Path, installer))
}

func (d *Dapp) removeFinalize() error {
	return nil
}

func copyServiceWrapper(d, s *Dapp) {
}
