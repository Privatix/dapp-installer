// +build windows

package util

import (
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/lxn/win"
	"github.com/privatix/dappctrl/util/log"
	"golang.org/x/sys/windows/registry"
)

// MinWindowsVersion is supported min windows version (Windows7 and newer)
const MinWindowsVersion byte = 6

// MinDBEngineVersion is supported min DB Engine version
const MinDBEngineVersion int = 9

// WinRegDBEngine64 is a registry path to 64 bit DBEngine installations
const WinRegDBEngine64 string = `SOFTWARE\PostgreSQL\Installations\`

// WinRegDBEngine32 is a registry path to 32 bit DBEngine installations
const WinRegDBEngine32 string = `SOFTWARE\WOW6432Node\PostgreSQL\Installations\`

// WinRegInstalledDapp is a registry path to dapp installations
const WinRegInstalledDapp string = `SOFTWARE\Privatix\Dapp\`

// WinRegUninstallDapp is a registry path to dapp uninstallations
const WinRegUninstallDapp string = `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\Privatix Dapp `

// CheckSystemPrerequisites does checked system to prerequisites.
func CheckSystemPrerequisites(volume string, logger log.Logger) bool {
	if runtime.GOOS != "windows" {
		logger.Warn("software install only to Windows platform")
		return false
	}

	if !checkWindowsVersion() {
		logger.Warn("Windows version does not meet the requirements")
		return false
	}

	if !checkMemory() {
		logger.Warn("RAM does not meet the requirements")
		return false
	}

	if !checkStorage(volume) {
		logger.Warn("available size of disk does not meet the requirements")
		return false
	}

	return true
}

func checkWindowsVersion() bool {
	v := win.GetVersion()
	major := byte(v)
	return major >= MinWindowsVersion
}

func checkStorage(volume string) bool {
	h := syscall.MustLoadDLL("kernel32.dll")
	c := h.MustFindProc("GetDiskFreeSpaceExW")
	lpFreeBytesAvailable := uint64(0)
	u := utf16.Encode([]rune(volume))
	c.Call(uintptr(unsafe.Pointer(&u[0])),
		uintptr(unsafe.Pointer(&lpFreeBytesAvailable)))
	return lpFreeBytesAvailable >= MinAvailableDiskSize
}

func checkMemory() bool {
	var totalMemoryInKilobytes uint64
	win.GetPhysicallyInstalledSystemMemory(&totalMemoryInKilobytes)
	return totalMemoryInKilobytes*1024 > MinMemorySize
}

// ExistingDBEnginePort returns existing db engine port number.
func ExistingDBEnginePort(logger log.Logger) (int, bool) {
	if p, ok := getDBEngineServicePort(WinRegDBEngine64); ok {
		return p, ok
	}
	return getDBEngineServicePort(WinRegDBEngine32)
}

func getDBEngineServicePort(keyPath string) (int, bool) {
	path, ok := checkDBEngineVersion(registry.LOCAL_MACHINE, keyPath)
	if ok {
		s := strings.Replace(keyPath, "Installations", "Services", 1)
		path = s + path
		return getServicePort(registry.LOCAL_MACHINE, path), true
	}
	return 0, false
}

// ExecuteCommand does executing file.
func ExecuteCommand(filename string, args []string) error {
	cmd := exec.Command(filename, args...)
	return cmd.Run()
}

// ExistingDapp returns info about existing dappctrl.
func ExistingDapp(role string, logger log.Logger) (map[string]string, bool) {
	maps, err := getInstalledDappVersion(role)
	if err != nil {
		return nil, false
	}
	return maps, len(maps) > 0
}

// RenamePath changes folder name and returns it
func RenamePath(path, folder string) string {
	path = strings.TrimSuffix(path, "\\")
	path = path[:strings.LastIndex(path, "\\")]

	return path + "\\" + folder + "\\"
}
