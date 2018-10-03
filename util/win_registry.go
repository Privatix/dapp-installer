// +build windows

package util

import (
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
