package tor

import (
	"path/filepath"

	"github.com/privatix/dapp-installer/unix"
	"github.com/privatix/dapp-installer/util"
)

// ServiceName returns tor daemon name.
func (t Tor) ServiceName() string {
	return "tor_" + util.Hash(t.RootPath)
}

func installService(daemon, path, descr, installUID string, autostart bool) error {
	t := filepath.Join(path, "tor", "tor")
	c := filepath.Join(path, "tor", "settings", "torrc")

	d := unix.NewDaemon(daemon)
	d.Command = t
	d.Args = []string{"-f", c}
	d.AutoStart = autostart
	d.UID = installUID

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

func removeService(daemon, installUID string) error {
	d := unix.NewDaemon(daemon)
	d.UID = installUID
	return d.Remove()
}
