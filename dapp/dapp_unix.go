// +build linux darwin

package dapp

import (
	"os"
	"path/filepath"

	"github.com/privatix/dapp-installer/unix"
	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dappctrl/util/log"
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

func (d *Dapp) configurateController(logger log.Logger) error {
	if err := d.modifyDappConfig(logger); err != nil {
		return err
	}

	logger.Info("create daemon")
	ctrl := d.Controller

	ctrl.Service.ID = d.controllerHash()

	ctrl.Service.Command = filepath.Join(d.InstallPath, ctrl.EntryPoint)
	ctrl.Service.Args = []string{
		"-config", filepath.Join(d.InstallPath, ctrl.Configuration)}

	if err := ctrl.Service.Install(); err != nil {
		logger.Warn("failed to install daemon:" + ctrl.Service.ID)
		return err
	}
	logger.Info("start daemon")
	if err := ctrl.Service.Start(); err != nil {
		// retry attempt to start service
		if err := ctrl.Service.Start(); err != nil {
			logger.Warn("failed to start daemon:" + ctrl.Service.ID)
			return err
		}
	}

	_, installer := filepath.Split(os.Args[0])
	return util.CopyFile(os.Args[0],
		filepath.Join(d.InstallPath, installer))
}

func (d *Dapp) prepareToInstall(logger log.Logger) error {
	return nil
}

func (d *Dapp) removeFinalize(logger log.Logger) error {
	return nil
}

func copyServiceWrapper(d, s *Dapp) {
}
