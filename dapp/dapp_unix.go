// +build linux darwin

package dapp

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/privatix/dapp-installer/dbengine"
	"github.com/privatix/dapp-installer/unix"
	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dappctrl/util/log"
)

type service struct {
	unix.Service
}

// NewConfig creates a default Dapp configuration.
func NewConfig() *Dapp {
	return &Dapp{
		DBEngine: dbengine.NewConfig(),
	}
}

// Exists returns existing dapp in the host.
func Exists(role string, logger log.Logger) (*Dapp, bool) {
	return nil, false
}

func (d *Dapp) configurateController(logger log.Logger) error {
	if err := d.modifyDappConfig(); err != nil {
		return err
	}

	_, installer := filepath.Split(os.Args[0])
	return util.CopyFile(os.Args[0],
		filepath.Join(d.InstallPath, installer))
}

func (d *Dapp) installFinalize(logger log.Logger) error {
	// TODO (ubozov) register dbengine service as deamon?
	return nil
}

func (d *Dapp) updateFinalize(logger log.Logger) error {
	return errors.New("not implemented")
}

func (d *Dapp) removeFinalize(logger log.Logger) error {
	return errors.New("not implemented")
}
