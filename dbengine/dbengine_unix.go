// +build linux darwin

package dbengine

import (
	"errors"
	"os/exec"
	"path/filepath"
)

func startService(installPath, user string) error {
	fileName := filepath.Join(installPath, `pgsql/bin/pg_ctl`)

	dataPath := filepath.Join(installPath, `pgsql/data`)
	cmd := exec.Command(fileName, "start", "-D", dataPath)

	return cmd.Run()
}

func removeService(installPath string) error {
	return errors.New("not implemented")
}

func stopService(installPath string) error {
	fileName := filepath.Join(installPath, `pgsql/bin/pg_ctl`)

	dataPath := filepath.Join(installPath, `pgsql/data`)
	cmd := exec.Command(fileName, "stop", "-D", dataPath)

	return cmd.Run()
}
