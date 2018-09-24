// +build windows

package util

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/lxn/win"
	"github.com/privatix/dappctrl/util/log"
)

// MinWindowsVersion is supported min windows version (Windows7 and newer)
const MinWindowsVersion byte = 6

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

// GrantAccess grants access to directory.
func GrantAccess(path, user string) error {
	cmd := exec.Command("icacls", path, "/t", "/grant",
		fmt.Sprintf("%s:F", user))
	return cmd.Run()
}
