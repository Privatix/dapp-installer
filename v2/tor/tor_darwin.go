package tor

import (
	"path/filepath"

	"github.com/privatix/dapp-installer/unix"
	"github.com/privatix/dapp-installer/util"
)

// ServiceName returns tor daemon name.
func (inst *Installation) ServiceName() string {
	return "tor_" + util.Hash(inst.RootPath)
}

func installService(daemon, path, descr string, autostart bool) error {
	t := filepath.Join(path, "tor", "tor")
	c := filepath.Join(path, "tor", "settings", "torrc")

	d := unix.NewDaemon(daemon)
	d.Command = t
	d.Args = []string{"-f", c}
	d.AutoStart = autostart

	return d.Install()
}

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

func removeService(daemon string) error {
	return unix.NewDaemon(daemon).Remove()
}
