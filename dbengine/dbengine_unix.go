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

func startService(installPath string) error {
	fileName := filepath.Join(installPath, `pgsql/bin/postgres`)
	dataPath := filepath.Join(installPath, `pgsql/data`)

	d := unix.NewDaemon(Hash(installPath))
	d.Command = fileName
	d.Args = []string{"-D" + dataPath}

	if err := d.Install(); err != nil {
		return err
	}
	return d.Start()
}

func removeService(installPath string) error {
	return unix.NewDaemon(Hash(installPath)).Remove()
}

func stopService(installPath string) error {
	return unix.NewDaemon(Hash(installPath)).Stop()
}
