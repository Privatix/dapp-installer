package dapp

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/rdegges/go-ipify"

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

func setDynamicPorts(configFile string) error {
	read, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	contents := string(read)
	addrs := util.MatchAddr(contents)
	reserves := make(map[string]bool)

	for _, addr := range addrs {
		port, err := util.FreePort(addr.Host, addr.Port)
		if err != nil {
			return err
		}

		// If the available port is already reserved,
		// the search for another unreserved available port in the loop.
		_, ok := reserves[port]
		for ok {
			p, _ := strconv.Atoi(port)
			port, _ := util.FreePort(addr.Host, strconv.Itoa(p+1))
			_, ok = reserves[port]
		}

		reserves[port] = true

		if port == addr.Port {
			continue
		}

		newAddress := strings.Replace(addr.Address,
			fmt.Sprintf(":%s", addr.Port),
			fmt.Sprintf(":%s", port), -1)
		contents = strings.Replace(contents, addr.Address,
			newAddress, -1)
	}
	return ioutil.WriteFile(configFile, []byte(contents), 0)
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
