// +build windows

package util

import (
	"strconv"
	"strings"

	"golang.org/x/sys/windows/registry"
)

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
