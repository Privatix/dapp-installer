package dapp

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	ipify "github.com/rdegges/go-ipify"

	"github.com/privatix/dapp-installer/util"
)

func setConfigurationValues(jsonMap map[string]interface{},
	settings map[string]interface{}) error {
	for key, value := range settings {
		path := strings.Split(key, ".")
		length := len(path) - 1
		m := jsonMap
		if length > 0 {
			for i := 0; i < length; i++ {
				item, ok := m[path[i]]
				if !ok || reflect.TypeOf(m) != reflect.TypeOf(item) {
					return fmt.Errorf("failed to set config params: %s", key)
				}
				m, _ = item.(map[string]interface{})
			}
		}
		m[path[length]] = value
	}
	return nil
}

// createFirewallRule creates firewall rule for payment reciever of dappctrl.
func createFirewallRule(d *Dapp, addr string) error {
	if runtime.GOOS != "windows" {
		return nil
	}

	s := strings.Split(addr, ":")
	if len(s) < 2 {
		return fmt.Errorf("failed to read payment reciever address")
	}
	args := []string{"-ExecutionPolicy", "Bypass",
		"-File",
		filepath.Join(d.Path, "dappctrl", "set-ctrlfirewall.ps1"),
		"-Create", "-ServiceName", d.Controller.Service.ID,
		"-ProgramPath",
		filepath.Join(d.Path, "dappctrl", "dappctrl.exe"),
		"-Port", s[1], "-Protocol", "tcp"}
	return util.ExecuteCommand("powershell", args...)
}

func externalIP() (string, error) {
	return ipify.GetIp()
}
