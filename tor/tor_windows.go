package tor

import (
	"fmt"
	"path/filepath"

	"github.com/privatix/dapp-installer/util"
)

// ServiceName returns tor service name.
func (t Tor) ServiceName() string {
	return "Privatix Tor " + util.Hash(t.RootPath)
}

func installService(service, path string) error {
	torExe := filepath.Join(path, `tor/tor`)
	torConf := filepath.Join(path, `tor/settings/torrc`)
	binPath := fmt.Sprintf(`"%s" -nt-service -f "%s"`, torExe, torConf)

	return util.ExecuteCommand("sc", "create", service, "start=auto",
		"binPath="+binPath)
}

func startService(service string) error {
	return util.ExecuteCommand("sc", "start", service)
}

func stopService(service string) error {
	return util.ExecuteCommand("sc", "stop", service)
}

func removeService(service string) error {
	return util.ExecuteCommand("sc", "delete", service)
}
