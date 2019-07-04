package flows

import (
	"path/filepath"
	"runtime"

	"github.com/privatix/dapp-installer/dapp"
	"github.com/privatix/dapp-installer/util"
)

func installSupervisorIfClient(d *dapp.Dapp) error {
	if d.Role != "client" {
		return nil
	}
	return runSupervisor(d, "install", "7777", d.Role, d.Path, d.Tor.RootPath)
}

func removeSupervisorIfClient(d *dapp.Dapp) error {
	if d.Role != "client" {
		return nil
	}
	return runSupervisor(d, "remove")
}

func runSupervisor(d *dapp.Dapp, args ...string) error {
	execPath, err := supervisorExecPath(d.Path)
	if err != nil {
		return err
	}

	return util.ExecuteCommand(execPath, args...)
}

func supervisorExecPath(dappRoot string) (string, error) {
	execName := "dapp-supervisor"
	if runtime.GOOS == "windows" {
		execName = "dapp-supervisor.exe"
	}
	return filepath.Abs(filepath.Join(dappRoot, "..", execName))
}
