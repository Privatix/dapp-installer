package tor

import (
	"path/filepath"

	"github.com/privatix/dapp-installer/util"
)

// ServiceName returns tor service name.
func (t Tor) ServiceName() string {
	return "Privatix Tor " + util.Hash(t.RootPath)
}

func installService(service, path, description string, autostart bool) error {
	torExe := filepath.Join(path, "tor", "tor")
	torConf := filepath.Join(path, "tor", "settings", "torrc")

	return util.CreateService(service, torExe, description, autostart,
		"-nt-service", "-f", torConf)
}

func startService(service, _ string) error {
	return util.StartService(service)
}

func stopService(service, _ string) error {
	return util.ExecuteCommand("sc", "stop", service)
}

func removeService(service string) error {
	return util.RemoveService(service)
}
