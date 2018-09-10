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
	Path string
	Keys []Key
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
func CreateRegistryKey(reg *Registry) error {
	key, _, err := registry.CreateKey(registry.LOCAL_MACHINE, reg.Path,
		registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer key.Close()

	for _, k := range reg.Keys {
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
func RemoveRegistryKey(reg *Registry) error {
	return registry.DeleteKey(registry.LOCAL_MACHINE, reg.Path)
}

func getInstalledDappVersion() (string, error) {
	k, err := getRegistryKeyByPath(registry.LOCAL_MACHINE, WinRegInstalledDapp)
	if err != nil {
		return "", err
	}
	defer k.Close()
	v, _, _ := k.GetStringValue("Version")
	return v, nil
}
