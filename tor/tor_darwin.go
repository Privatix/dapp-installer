package tor

import (
	"path/filepath"

	"github.com/privatix/dapp-installer/unix"
	"github.com/privatix/dapp-installer/util"
)

func (t Tor) serviceName() string {
	return "tor_" + util.Hash(t.RootPath)
}

func installService(daemon, path string) error {
	t := filepath.Join(path, `tor/tor`)
	c := filepath.Join(path, `tor/settings/torrc`)

	d := unix.NewDaemon(daemon)
	d.Command = t
	d.Args = []string{"-f", c}

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
