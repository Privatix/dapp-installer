package flows

import (
	"fmt"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/privatix/dapp-installer/dapp"
	"github.com/privatix/dapp-installer/util"
)

func installSupervisorIfClient(d *dapp.Dapp) error {
	if d.Role != "client" {
		return nil
	}

	args := []string{"install", "7777", d.Role, d.Path, d.Tor.RootPath}
	if runtime.GOOS == "darwin" {
		user, err := user.Current()
		if err != nil {
			return err
		}
		// On darwin some services run under current user, not root.
		// Passing supervisor the uid of user so it cans run/stop services.
		args = append(args, user.Uid)
	}
	return runSupervisor(d, args...)
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

	if runtime.GOOS == "darwin" {
		txt := `with prompt "Privatix wants to make changes"`
		evelate := "with administrator privileges"
		command := fmt.Sprintf("%s %s", execPath, strings.Join(args, " "))
		script := fmt.Sprintf(`do shell script "sudo %s" %s %s`,
			command, txt, evelate)

		return util.ExecuteCommand("osascript", "-e", script)
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
