package command

import (
	"fmt"
	"path/filepath"

	"github.com/privatix/dapp-installer/container"
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

func getContainer(d *dapp.Dapp) *container.Container {
	c := container.NewContainer()

	c.Name = d.Role
	c.Path = d.Path

	return c
}

func installContainer(d *dapp.Dapp) error {
	c := getContainer(d)

	if err := c.Install(); err != nil {
		return fmt.Errorf("failed to install container: %v", err)
	}

	return nil
}

func startContainer(d *dapp.Dapp) error {
	c := getContainer(d)

	if err := c.Start(); err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	return nil
}

func stopContainer(d *dapp.Dapp) error {
	c := getContainer(d)

	if err := c.Stop(); err != nil {
		return fmt.Errorf("failed to stop container: %v", err)
	}

	return nil
}

func removeContainer(d *dapp.Dapp) error {
	c := getContainer(d)

	if err := c.Remove(); err != nil {
		return fmt.Errorf("failed to remove container: %v", err)
	}

	return nil
}

func configure(d *dapp.Dapp) error {
	d.Tor.HiddenServiceDir = "var/lib/tor/hidden_service"
	d.Tor.Config = "torrcprod"
	d.Tor.SocksPort = 9099
	d.Tor.IsLinux = true

	if err := d.Tor.Install(d.Role); err != nil {
		return fmt.Errorf("failed to install tor: %v", err)
	}

	path := filepath.Join(d.Path, d.Controller.Configuration)
	err := util.SetDynamicPorts(path)
	if err != nil {
		return fmt.Errorf("failed to set dynamic ports: %v", err)
	}

	return nil
}

func finalize(d *dapp.Dapp) error {

	return nil
}
