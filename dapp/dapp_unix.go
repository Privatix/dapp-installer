// +build linux darwin

package dapp

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/privatix/dapp-installer/data"
	"github.com/privatix/dappctrl/util/log"
)

// Install installs a dapp core.
func (d *Dapp) Install(db *data.DB, logger log.Logger, ok bool) error {
	logger.Warn(fmt.Sprintf("your OS %s is not supported now", runtime.GOOS))
	return errors.New("is not supported for this feature")
}

// Remove removes installed dapp core.
func (d *Dapp) Remove() error {
	logger.Warn(fmt.Sprintf("your OS %s is not supported now", runtime.GOOS))
	return errors.New("is not supported for this feature")
}
