package flows

import (
	"context"
	"fmt"

	"github.com/privatix/dapp-installer/dapp"
)

func installTor(d *dapp.Dapp) error {
	autostart := d.Role == "agent"
	if err := d.Tor.Install(d.Role, autostart, d.UID); err != nil {
		return fmt.Errorf("failed to install tor: %v", err)
	}
	return nil
}

func removeTor(d *dapp.Dapp) error {
	if err := d.Tor.Remove(d.UID); err != nil {
		return fmt.Errorf("failed to remove tor: %v", err)
	}

	return nil
}

func stopTor(d *dapp.Dapp) error {
	if err := d.Tor.Stop(context.Background(), d.UID); err != nil {
		return fmt.Errorf("failed to stop tor: %v", err)
	}

	return nil
}

func startTor(d *dapp.Dapp) error {
	if err := d.Tor.Start(context.Background(), d.UID); err != nil {
		return fmt.Errorf("failed to start tor: %v", err)
	}

	return nil
}
