// +build linux darwin

package util

// #include <unistd.h>
import "C"

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"github.com/privatix/dapp-installer/unix"
)

// CheckSystemPrerequisites does checked system to prerequisites.
func CheckSystemPrerequisites(volume string) error {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		return fmt.Errorf("software install only to unix based platform")
	}

	// TODO (ubozov) check platform version?

	if !checkMemory() {
		return fmt.Errorf("RAM does not meet the requirements")
	}

	if !checkStorage() {
		return fmt.Errorf("available size of disk does not meet the requirements")
	}
	return nil
}

func checkStorage() bool {
	var stat syscall.Statfs_t
	wd, _ := os.Getwd()
	syscall.Statfs(wd, &stat)

	return stat.Bavail*uint64(stat.Bsize) > MinAvailableDiskSize
}

func checkMemory() bool {
	memSize := C.sysconf(C._SC_PHYS_PAGES) * C.sysconf(C._SC_PAGE_SIZE)
	return uint64(memSize) > MinMemorySize
}

// GrantAccess grants access to directory.
func GrantAccess(path string) error {
	return os.Chown(path, os.Getuid(), os.Getgid())
}

// IsServiceStopped returns service stopped status.
func IsServiceStopped(id, uid string) bool {
	d := unix.NewDaemon(id)
	d.UID = uid
	return d.IsStopped()
}

// ExecuteCommand executes a file.
func ExecuteCommand(filename string, args ...string) (err error) {
	cmd := exec.Command(filename, args...)
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
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return output.String(), nil
}
