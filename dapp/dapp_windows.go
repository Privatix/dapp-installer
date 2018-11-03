// +build windows

package dapp

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dapp-installer/windows"
)

type service struct {
	windows.Service
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

func (d *Dapp) controllerHash() string {
	return fmt.Sprintf("Privatix Controller %s", util.Hash(d.Path))
}

func (d *Dapp) createSymlink() {
	target := filepath.Join(d.Path, d.Gui.EntryPoint)
	linkName := path.Base(d.Gui.EntryPoint)
	link := filepath.Join(util.DesktopPath(),
		fmt.Sprintf(`%s %s`, linkName, d.Role))

	if len(filepath.Ext(linkName)) == 0 {
		target += ".exe"
	}

	exec.Command("cmd", "/C", "mklink", link, target).Run()
}

// Configurate configurates installed dapp.
func (d *Dapp) Configurate() error {
	if d.Gui.Symlink {
		d.createSymlink()
	}

	if err := d.modifyDappConfig(); err != nil {
		return err
	}

	_, installer := filepath.Split(os.Args[0])
	util.CopyFile(os.Args[0], filepath.Join(d.Path, installer))

	ctrl := d.Controller
	ctrlPath := filepath.Join(d.Path, filepath.Dir(ctrl.EntryPoint))

	hash := d.controllerHash()
	ctrl.Service.ID = hash
	ctrl.Service.Name = hash
	ctrl.Service.Description = fmt.Sprintf("dapp controller %s", hash)
	ctrl.Service.GUID = filepath.Join(ctrlPath, ctrl.Service.ID)
	if err := ctrl.Service.CreateWrapper(ctrlPath); err != nil {
		return fmt.Errorf("failed to create service wrapper: %v", err)
	}

	if err := ctrl.Service.Install(); err != nil {
		return fmt.Errorf("failed to install service: %v", err)
	}

	if err := ctrl.Service.Start(); err != nil {
		// retry attempt to start service
		if err := ctrl.Service.Start(); err != nil {
			return fmt.Errorf("failed to start service: %v", err)
		}
	}

	return nil
}

func copyServiceWrapper(d, s *Dapp) {
	hash := s.controllerHash()
	scvExe := "dappctrl/" + hash + ".exe"
	scvConfig := "dappctrl/" + hash + ".config.json"

	util.CopyFile(filepath.Join(s.Path, scvExe),
		filepath.Join(d.Path, scvExe))
	util.CopyFile(filepath.Join(s.Path, scvConfig),
		filepath.Join(d.Path, scvConfig))
}
