// +build windows

package util

import (
	"runtime"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/Privatix/dappctrl/util/log"
	"github.com/lxn/win"
)

// MinWindowsVersion is supported min windows version (Windows7 and newer)
const MinWindowsVersion byte = 6

// CheckSystemPrerequisites does checked system to prerequisites
func CheckSystemPrerequisites(log log.Logger) bool {
	if runtime.GOOS != "windows" {
		log.Warn("software install only to Windows platform")
		return false
	}

	if !checkWindowsVersion() {
		log.Warn("Windows version does not meet the requirements")
		return false
	}

	if !checkMemory() {
		log.Warn("RAM does not meet the requirements")
		return false
	}

	if !checkStorage() {
		log.Warn("available size of disk does not meet the requirements")
		return false
	}

	return true
}

func checkWindowsVersion() bool {
	v := win.GetVersion()
	major := byte(v)
	return major >= MinWindowsVersion
}

func checkStorage() bool {
	h := syscall.MustLoadDLL("kernel32.dll")
	c := h.MustFindProc("GetDiskFreeSpaceExW")
	lpFreeBytesAvailable := uint64(0)
	u := utf16.Encode([]rune("\\"))
	c.Call(uintptr(unsafe.Pointer(&u[0])),
		uintptr(unsafe.Pointer(&lpFreeBytesAvailable)))

	return lpFreeBytesAvailable >= MinAvailableDiskSize
}

func checkMemory() bool {
	var totalMemoryInKilobytes uint64
	win.GetPhysicallyInstalledSystemMemory(&totalMemoryInKilobytes)
	return totalMemoryInKilobytes*1024 > MinMemorySize
}
