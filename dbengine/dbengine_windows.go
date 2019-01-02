// +build windows

package dbengine

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/privatix/dapp-installer/util"
)

// Hash returns db engine service unique ID.
func Hash(path string) string {
	return fmt.Sprintf("Privatix DB %s", util.Hash(path))
}

func runService(service string) error {
	out, err := util.ExecuteCommandOutput("sc", "queryex", service)
	if err != nil {
		return err
	}

	// service is running
	if strings.Contains(out, "RUNNING") {
		return nil
	}

	// trying start service
	return util.ExecuteCommand("net", "start", service)
}

func startService(installPath string) error {
	fileName := filepath.Join(installPath, "pgsql", "bin", "pg_ctl")
	serviceName := Hash(installPath)

	_ = util.ExecuteCommand(fileName, "unregister", "-N", serviceName)

	dataPath := filepath.Join(installPath, "pgsql", "data")
	err := util.ExecuteCommand(fileName, "register",
		"-N", serviceName, "-D", dataPath)

	if err != nil {
		return err
	}

	return runService(serviceName)
}

func removeService(installPath string) error {
	serviceName := Hash(installPath)
	stopService(installPath)

	fileName := filepath.Join(installPath, "pgsql", "bin", "pg_ctl")

	return util.ExecuteCommand(fileName, "unregister", "-N", serviceName)
}

func stopService(installPath string) error {
	serviceName := Hash(installPath)
	return util.ExecuteCommand("sc", "stop", serviceName)
}

func prepareToInstall(installPath string) error {
	// installs run-time components (the Visual C++ Redistributable Packages
	// for VS 2013) that are required to run postgresql database engine.
	if util.CheckInstallVCRedist() {
		return nil
	}
	vcredist := filepath.Join(installPath, "util", "vcredist_x64.exe")
	return util.ExecuteCommand(vcredist, "/install", "/quiet", "/norestart")
}
