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

func installService(daemon, path, descr string, autostart bool) error {
	t := filepath.Join(path, "tor", "tor")
	c := filepath.Join(path, "tor", "settings", "torrc")

	d := unix.NewDaemon(daemon)
	d.Command = t
	d.Args = []string{"-f", c}
	d.AutoStart = autostart

	return d.Install()
}

func startService(daemon string) error {
	return unix.NewDaemon(daemon).Start()
}

func stopService(daemon string) error {
	return unix.NewDaemon(daemon).Stop()
}

func removeService(daemon string) error {
	return unix.NewDaemon(daemon).Remove()
}
