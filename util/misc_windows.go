// +build windows

package util

import (
	"bytes"
	"fmt"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

// MinWindowsVersion is supported min windows version (Windows7 and newer)
const MinWindowsVersion byte = 6

// CheckSystemPrerequisites does checked system to prerequisites.
func CheckSystemPrerequisites(volume string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("software install only to Windows platform")
	}

	if !checkWindowsVersion() {
		return fmt.Errorf("Windows version does not meet the requirements")
	}

	if !checkMemory() {
		return fmt.Errorf("RAM does not meet the requirements")
	}

	if !checkStorage(volume) {
		return fmt.Errorf("available size of disk does not meet the requirements")
	}

	return nil
}

func checkWindowsVersion() bool {
	h := syscall.MustLoadDLL("kernel32.dll")
	c := h.MustFindProc("GetVersion")
	r, _, _ := c.Call()
	return byte(r) >= MinWindowsVersion
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
	h := syscall.MustLoadDLL("kernel32.dll")
	c := h.MustFindProc("GlobalMemoryStatusEx")

	type memoryStatus struct {
		Length               uint32
		MemoryLoad           uint32
		TotalPhys            uint64
		AvailPhys            uint64
		TotalPageFile        uint64
		AvailPageFile        uint64
		TotalVirtual         uint64
		AvailVirtual         uint64
		AvailExtendedVirtual uint64
	}

	var buf memoryStatus
	buf.Length = uint32(unsafe.Sizeof(buf))
	c.Call(uintptr(unsafe.Pointer(&buf)))
	return buf.TotalPhys >= MinMemorySize
}

// GrantAccess grants access to directory.
func GrantAccess(path string) error {
	u, err := user.Current()
	if err != nil {
		return err
	}
	cmd := exec.Command("icacls", path, "/t", "/grant",
		fmt.Sprintf("%s:F", u.Username))
	return cmd.Run()
}

// IsServiceStopped returns service stopped status.
func IsServiceStopped(service string) bool {
	checkServiceCmd := exec.Command("sc", "queryex", service)
	var checkServiceStdOut bytes.Buffer
	checkServiceCmd.Stdout = &checkServiceStdOut
	if err := checkServiceCmd.Run(); err != nil {
		return false
	}
	return strings.Contains(checkServiceStdOut.String(), "STOPPED")
}
