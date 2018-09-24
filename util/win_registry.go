// +build windows

package util

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// Key is a windows registry key.
type Key struct {
	Name  string
	Type  string
	Value string
}

// Registry is a windows registry path with contains keys.
type Registry struct {
	Install   []Key
	Uninstall []Key
}

func getRegistryKeyByPath(key registry.Key, p string) (*registry.Key, error) {
	k, err := registry.OpenKey(key, p,
		registry.QUERY_VALUE|registry.ENUMERATE_SUB_KEYS)

	if err != nil {
		return nil, err
	}
	return &k, nil
}

func getServicePort(key registry.Key, path string) int {
	k, err := getRegistryKeyByPath(key, path)
	if err != nil {
		return 0
	}
	defer k.Close()
	v, _, _ := k.GetIntegerValue("Port")
	return int(v)
}

func checkDBEngineVersion(key registry.Key, path string) (string, bool) {
	k, err := getRegistryKeyByPath(key, path)
	if err != nil {
		return "", false
	}
	defer k.Close()
	s, _, _ := k.GetStringValue("Version")
	if len(s) > 0 {
		major, _ := strconv.Atoi(s[0:strings.Index(s, ".")])
		if major >= MinDBEngineVersion {
			return path, true
		}
	}
	subNames, err := k.ReadSubKeyNames(-1)
	if err != nil {
		return "", false
	}
	for _, each := range subNames {
		if p, ok := checkDBEngineVersion(*k, each); ok {
			return p, true
		}
	}
	return "", false
}

// CreateRegistryKey creates new registry key in windows registry.
func CreateRegistryKey(reg *Registry, role string) error {
	err := createRegistryKey(WinRegInstalledDapp+role, reg.Install)
	if err != nil {
		return err
	}
	return createRegistryKey(WinRegUninstallDapp+"("+role+")", reg.Uninstall)
}

func createRegistryKey(path string, keys []Key) error {
	key, _, err := registry.CreateKey(registry.LOCAL_MACHINE, path,
		registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer key.Close()

	for _, k := range keys {
		switch k.Type {
		case "string":
			if err := key.SetStringValue(k.Name, k.Value); err != nil {
				return err
			}
		case "dword":
			v, _ := strconv.Atoi(k.Value)
			if err := key.SetDWordValue(k.Name, uint32(v)); err != nil {
				return err
			}
		default:
			return fmt.Errorf("creating registry key with type %s not implemented", k.Type)
		}
	}
	return nil
}

// RemoveRegistryKey removes registry key from windows registry.
func RemoveRegistryKey(reg *Registry, version string) error {
	err := registry.DeleteKey(registry.LOCAL_MACHINE,
		WinRegInstalledDapp+version)
	if err != nil {
		return err
	}
	return registry.DeleteKey(registry.LOCAL_MACHINE,
		WinRegUninstallDapp+version)
}

func getInstalledDappVersion(role string) (map[string]string, error) {
	root, err := getRegistryKeyByPath(registry.LOCAL_MACHINE, WinRegInstalledDapp)
	if err != nil {
		return nil, err
	}
	defer root.Close()

	subNames, err := root.ReadSubKeyNames(-1)
	if err != nil {
		return nil, err
	}
	for _, each := range subNames {
		if strings.EqualFold(role, each) {
			k, err := getRegistryKeyByPath(*root, each)
			if err != nil {
				return nil, err
			}
			defer k.Close()

			v, _, _ := k.GetStringValue("Version")
			s, _, _ := k.GetStringValue("ServiceID")
			p, _, _ := k.GetStringValue("BaseDirectory")
			c, _, _ := k.GetStringValue("Configuration")
			g, _, _ := k.GetStringValue("Gui")
			sh, _, _ := k.GetStringValue("Shortcuts")

			maps := make(map[string]string)
			maps["Version"] = v
			maps["ServiceID"] = s
			maps["BaseDirectory"] = p
			maps["Configuration"] = c
			maps["Gui"] = g
			maps["Shortcuts"] = sh
			return maps, nil
		}
	}

	return nil, nil
}

// DesktopPath returns windows desktop path.
func DesktopPath() string {
	key, err := getRegistryKeyByPath(registry.LOCAL_MACHINE,
		`Software\Microsoft\Windows\CurrentVersion\Explorer\Shell Folders`)
	if err != nil {
		return ""
	}
	defer key.Close()

	val, _, _ := key.GetStringValue("Common Desktop")
	return val
}

// ProgramFilesPath returns windows program files path.
func ProgramFilesPath() string {
	key, err := getRegistryKeyByPath(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows\CurrentVersion`)
	if err != nil {
		return ""
	}
	defer key.Close()

	val, _, _ := key.GetStringValue("ProgramFilesDir")
	return val
}
