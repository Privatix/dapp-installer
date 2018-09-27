// +build linux darwin

package dbengine

import (
	"errors"
)

func startService(installPath, user string, envs []string) error {
	return errors.New("not implemented")
}

func removeService(installPath string) error {
	return errors.New("not implemented")
}

func stopService(installPath string) error {
	return errors.New("not implemented")
}
