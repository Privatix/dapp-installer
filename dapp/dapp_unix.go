// +build linux darwin

package dapp

import (
	"fmt"
	"os"
	"path"
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

func (d *Dapp) controllerHash() string {
	return fmt.Sprintf("ctrl_%s", util.Hash(d.Path))
}
func (d *Dapp) createSymlink() {
	target := filepath.Join(d.Path, d.Gui.EntryPoint)
	linkName := path.Base(d.Gui.EntryPoint)
	link := filepath.Join(util.DesktopPath(),
		fmt.Sprintf("%s %s", linkName, d.Role))

	if len(filepath.Ext(linkName)) == 0 {
		target += ".app"
	}

	os.Symlink(target, link)
}

// Configurate configurates installed dapp.
func (d *Dapp) Configurate() error {
	if d.Gui.Symlink {
		d.createSymlink()
	}

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

func copyServiceWrapper(d, s *Dapp) {
}
