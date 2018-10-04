// +build linux darwin

package dapp

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/privatix/dapp-installer/unix"
	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dappctrl/util/log"
)

func newService() *service {
	return &service{
		unix.Service{
			Command: "dappctrl",
			Args:    []string{"-config", "dappctrl.config.json"},
		},
	}
}

func (d *Dapp) configurateController(logger log.Logger) error {
	if err := d.modifyDappConfig(logger); err != nil {
		return err
	}

	_, installer := filepath.Split(os.Args[0])
	return util.CopyFile(os.Args[0],
		filepath.Join(d.InstallPath, installer))
}

func (d *Dapp) prepareToInstall(logger log.Logger) error {
	return nil
}

func (d *Dapp) removeFinalize(logger log.Logger) error {
	return errors.New("not implemented")
}

func copyServiceWrapper(d, s *Dapp) {
	return errors.New("not implemented")
}
