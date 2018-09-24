// +build linux darwin

package dapp

import (
	"errors"

	"github.com/privatix/dappctrl/util/log"
)

func (d *Dapp) configurateController(logger log.Logger) error {
	return errors.New("does not implemented")
}

func (d *Dapp) installFinalize(logger log.Logger) error {
	return errors.New("does not implemented")
}

func (d *Dapp) updateFinalize(logger log.Logger) error {
	return errors.New("does not implemented")
}

func (d *Dapp) removeFinalize(logger log.Logger) error {
	return errors.New("does not implemented")
}
