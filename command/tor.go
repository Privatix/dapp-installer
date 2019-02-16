package command

import (
	"fmt"

	"github.com/privatix/dapp-installer/dapp"
)

func installTor(d *dapp.Dapp) error {
	if err := d.Tor.Install(d.Role); err != nil {
		return fmt.Errorf("failed to install tor: %v", err)
	}
	return nil
}

func removeTor(d *dapp.Dapp) error {
	if err := d.Tor.Remove(); err != nil {
		return fmt.Errorf("failed to remove tor: %v", err)
	}

	return nil
}

func stopTor(d *dapp.Dapp) error {
	if err := d.Tor.Stop(); err != nil {
		return fmt.Errorf("failed to stop tor: %v", err)
	}

	return nil
}

func startTor(d *dapp.Dapp) error {
	if err := d.Tor.Start(); err != nil {
		return fmt.Errorf("failed to start tor: %v", err)
	}

	return nil
}
