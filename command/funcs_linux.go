package command

import (
	"fmt"

	"github.com/privatix/dapp-installer/dapp"
	"github.com/privatix/dapp-installer/util"
)

func prepare(d *dapp.Dapp) error {
	if err := util.ExecuteCommand("apt-get", "update"); err != nil {
		return fmt.Errorf("failed to prepare system: %v", err)
	}

	if err := util.ExecuteCommand("apt-get", "install",
		"systemd-container", "-y"); err != nil {
		return fmt.Errorf("failed to prepare system: %v", err)
	}

	if err := util.ExecuteCommand("apt-get", "install", "lshw",
		"-y"); err != nil {
		return fmt.Errorf("failed to prepare system: %v", err)
	}

	return nil
}
