// +build linux darwin

package dbengine

import (
	"fmt"
	"path/filepath"

	"github.com/privatix/dapp-installer/unix"
	"github.com/privatix/dapp-installer/util"
)

// Hash returns db engine service unique ID.
func Hash(installPath string) string {
	return fmt.Sprintf("db_%s", util.Hash(installPath))
}

func startService(installPath, installUID string, autostart bool) error {
	fileName, err := filepath.Abs(filepath.Join(installPath, "pgsql", "bin", "postgres"))
	if err != nil {
		return err
	}
	dataPath, err := filepath.Abs(filepath.Join(installPath, "pgsql", "data"))

	d := unix.NewDaemon(Hash(installPath))
	d.Command = fileName
	d.Args = []string{"-D" + dataPath}
	d.AutoStart = autostart
	d.UID = installUID

	if err := d.Install(); err != nil {
		return err
	}
	return d.Start()
}

func removeService(installPath string) error {
	return unix.NewDaemon(Hash(installPath)).Remove()
}

func stopService(installPath, installUID string) error {
	d := unix.NewDaemon(Hash(installPath))
	d.UID = installUID
	return d.Stop()
}

func prepareToInstall(installPath string) error {
	return nil
}
