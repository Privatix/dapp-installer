package flows

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/privatix/dapp-installer/container"
	"github.com/privatix/dapp-installer/dapp"
	"github.com/privatix/dapp-installer/data"
	"github.com/privatix/dapp-installer/product"
	"github.com/privatix/dapp-installer/util"
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

	version, ok := data.ReadAppVersion(d.DBEngine.DB)
	if !ok {
		return fmt.Errorf("failed to read app version")
	}

	d.Version = version

	return nil
}

func updateContainer(d *dapp.Dapp) error {
	if strings.HasSuffix(d.Path, "/") {
		d.Path = d.Path[:len(d.Path)-1]
	}

	oldDapp := *d

	// Extract.
	b, dir := filepath.Split(d.Path)
	newPath := filepath.Join(b, dir+"_new")
	d.Path = newPath
	if err := extractAndUpdateVersion(d); err != nil {
		d.Path = oldDapp.Path
		return err
	}
	d.Path = oldDapp.Path

	defer os.RemoveAll(newPath)

	// Read and store new version value.
	version := util.ParseVersion(d.Version)
	if util.ParseVersion(oldDapp.Version) >= version {
		return fmt.Errorf(
			"dapp current version: %s, update is not required",
			oldDapp.Version)
	}

	if err := product.Update(d.Role, oldDapp.Path, newPath,
		d.Product); err != nil {
		return fmt.Errorf("failed to update products: %v", err)
	}

	c := getContainer(d)
	copies := []string{"var/lib/postgresql/10/main",
		"var/lib/tor", "etc/tor", "etc/systemd/system"}

	merges := []string{d.Controller.Configuration, d.Gui.Configuration}
	d.BackupPath = filepath.Join(b, dir+"_backup")
	if err := c.Update(newPath, d.BackupPath, copies, merges); err != nil {
		return fmt.Errorf("failed to update dapp: %v", err)
	}

	return configureDapp(d)
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
