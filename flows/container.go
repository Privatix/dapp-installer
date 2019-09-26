package flows

import (
	"fmt"
	"os"

	"github.com/privatix/dapp-installer/container"
	"github.com/privatix/dapp-installer/dapp"
)

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

func enableAndStartContainer(d *dapp.Dapp) error {
	c := getContainer(d)

	if err := c.EnableAndStart(); err != nil {
		return fmt.Errorf("failed to enable and start container: %v", err)
	}

	return nil
}

func disableAndStopContainer(d *dapp.Dapp) error {
	c := getContainer(d)

	if err := c.DisableAndStop(); err != nil {
		return fmt.Errorf("failed to disable and stop container: %v", err)
	}

	return nil
}

func disablContainerIfClient(d *dapp.Dapp) error {
	if d.Role != "client" {
		return nil
	}
	c := getContainer(d)

	if err := c.Disable(); err != nil {
		return fmt.Errorf("failed to disable container: %v", err)
	}

	return nil
}

func restartContainer(d *dapp.Dapp) error {
	c := getContainer(d)

	if err := c.Restart(); err != nil {
		return fmt.Errorf("failed to restart container: %v", err)
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

func checkContainer(d *dapp.Dapp) error {
	if err := d.ReadConfig(); err != nil {
		return fmt.Errorf("failed to read config: %v", err)
	}

	if err := d.DBEngine.Ping(); err != nil {
		return fmt.Errorf("failed to check contianer: %v", err)
	}

	return nil
}

func restoreContainer(d *dapp.Dapp) error {
	if len(d.BackupPath) == 0 {
		return nil
	}

	if err := os.RemoveAll(d.Path); err != nil {
		return err
	}

	return os.Rename(d.BackupPath, d.Path)
}
