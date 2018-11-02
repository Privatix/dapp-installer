// +build windows

package dbengine

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/privatix/dapp-installer/util"
)

// Hash returns db engine service unique ID.
func Hash(path string) string {
	return fmt.Sprintf("Privatix DB %s", util.Hash(path))
}

func runService(service string) error {
	checkServiceCmd := exec.Command("sc", "queryex", service)

	var checkServiceStdOut bytes.Buffer
	checkServiceCmd.Stdout = &checkServiceStdOut

	if err := checkServiceCmd.Run(); err != nil {
		return err
	}

	// service is running
	if strings.Contains(checkServiceStdOut.String(), "RUNNING") {
		return nil
	}

	// trying start service
	return exec.Command("net", "start", service).Run()
}

func startService(installPath string) error {
	fileName := filepath.Join(installPath, `pgsql/bin/pg_ctl`)
	serviceName := Hash(installPath)

	exec.Command(fileName, "unregister", "-N", serviceName).Run()

	dataPath := filepath.Join(installPath, `pgsql/data`)
	cmd := exec.Command(fileName, "register",
		"-N", serviceName, "-D", dataPath)

	if err := cmd.Run(); err != nil {
		return err
	}

	return runService(serviceName)
}

func removeService(installPath string) error {
	serviceName := Hash(installPath)
	stopService(installPath)

	fileName := filepath.Join(installPath, `pgsql/bin/pg_ctl`)

	return exec.Command(fileName, "unregister", "-N", serviceName).Run()
}

func stopService(installPath string) error {
	serviceName := Hash(installPath)
	checkServiceCmd := exec.Command("sc", "stop", serviceName)

	return checkServiceCmd.Run()
}
