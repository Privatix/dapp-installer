// +build linux darwin

package util

// #include <unistd.h>
import "C"

import (
	"os"
	"runtime"
	"syscall"

	"github.com/privatix/dapp-installer/unix"
	"github.com/privatix/dappctrl/util/log"
)

// CheckSystemPrerequisites does checked system to prerequisites.
func CheckSystemPrerequisites(volume string, logger log.Logger) bool {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		logger.Warn("software install only to unix based platform")
		return false
	}

	// TODO (ubozov) check platform version?

	if !checkMemory() {
		logger.Warn("RAM does not meet the requirements")
		return false
	}

	if !checkStorage() {
		logger.Warn("available size of disk does not meet the requirements")
		return false
	}
	return true
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
