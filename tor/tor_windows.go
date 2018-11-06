package tor

import (
	"fmt"
	"os/exec"
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
	binPath := fmt.Sprintf("%s -nt-service -f %s", torExe, torConf)

	return exec.Command("sc", "create", service, "start=auto",
		"binPath="+binPath).Run()
}

func startService(service string) error {
	return exec.Command("sc", "start", service).Run()
}

func stopService(service string) error {
	return exec.Command("sc", "stop", service).Run()
}

func removeService(service string) error {
	return exec.Command("sc", "delete", service).Run()
}
