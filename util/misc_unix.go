// +build linux darwin

package util

// #include <unistd.h>
import "C"

import (
	"os"
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
func IsServiceStopped(id string) bool {
	return unix.NewDaemon(id).IsStopped()
}
