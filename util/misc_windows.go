// +build windows

package util

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc/mgr"
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
	return ExecuteCommand("icacls", path, "/t", "/grant",
		fmt.Sprintf("%s:F", u.Username))
}

// IsServiceStopped returns service stopped status.
func IsServiceStopped(service, _ string) bool {
	out, err := ExecuteCommandOutput("sc", "queryex", service)
	if err != nil {
		return false
	}
	return strings.Contains(out, "STOPPED")
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, f.Mode()); err != nil {
				return err
			}
		} else {
			if err := extractFile(fpath, rc, f); err != nil {
				return err
			}
		}
	}
	return nil
}

func extractFile(fpath string, rc io.ReadCloser, f *zip.File) error {
	err := os.MkdirAll(filepath.Dir(fpath), f.Mode())
	if err != nil {
		return err
	}
	outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		f.Mode())
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, rc)

	return err
}

// ExecuteCommand executes a file.
func ExecuteCommand(filename string, args ...string) (err error) {
	cmd := exec.Command(filename, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	if err = cmd.Run(); err != nil {
		outStr, errStr := outbuf.String(), errbuf.String()
		err = fmt.Errorf("%v\nout:\n%s\nerr:\n%s", err, outStr, errStr)
	}
	return err
}

// ExecuteCommandOutput does executing file and returns output.
func ExecuteCommandOutput(filename string, args ...string) (string, error) {
	cmd := exec.Command(filename, args...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return output.String(), nil
}

// CreateService creates the windows service.
func CreateService(name, exe, descr string, autostart bool, args ...string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	startType := mgr.StartAutomatic
	if autostart == false {
		startType = mgr.StartManual
	}

	s, err := m.CreateService(name, exe,
		mgr.Config{
			DisplayName: name,
			Description: descr,
			StartType: uint32(startType),
		}, args...)
	if err != nil {
		return err
	}
	defer s.Close()

	recoveryActions := []mgr.RecoveryAction{
		mgr.RecoveryAction{Type: mgr.ServiceRestart, Delay: 1000},
		mgr.RecoveryAction{Type: mgr.ServiceRestart, Delay: 2000},
		mgr.RecoveryAction{Type: mgr.ServiceRestart, Delay: 5000},
	}

	return s.SetRecoveryActions(recoveryActions, 0)
}

// RemoveService removes the windows service.
func RemoveService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()

	return s.Delete()
}

// StartService starts the windows service.
func StartService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()

	return s.Start()
}

// CheckInstallVCRedist checks to install Visual C++ Redistributable Packages.
func CheckInstallVCRedist() bool {
	value, err := getWinRegistryBoolValue(
		`SOFTWARE\Microsoft\VisualStudio\12.0\VC\Runtimes\x64`,
		"Installed")
	if value && err == nil {
		return true
	}

	value, err = getWinRegistryBoolValue(
		`SOFTWARE\WOW6432Node\Microsoft\VisualStudio\12.0\VC\Runtimes\x64`,
		"Installed")

	return value && err == nil
}

func getWinRegistryBoolValue(regPath, parameter string) (bool, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, regPath,
		registry.QUERY_VALUE)

	if err != nil {
		return false, err
	}

	defer key.Close()

	value, _, err := key.GetIntegerValue(parameter)
	if err != nil {
		return false, err
	}

	return value == 1, nil
}
