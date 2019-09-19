package service

import (
	"github.com/privatix/dapp-installer/unix"
)

func startService(daemon, installUID string) error {
	d := unix.NewDaemon(daemon)
	d.UID = installUID
	return d.Start()
}

func stopService(daemon, installUID string) error {
	d := unix.NewDaemon(daemon)
	d.UID = installUID
	return d.Stop()
}
