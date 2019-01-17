package tor

import (
	"github.com/privatix/dapp-installer/util"
)

// ServiceName returns tor daemon name.
func (t Tor) ServiceName() string {
	return "tor_" + util.Hash(t.RootPath)
}

func installService(daemon, path, descr string) error {
	return nil
}

func startService(daemon string) error {
	return nil
}

func stopService(daemon string) error {
	return nil
}

func removeService(daemon string) error {
	return nil
}
