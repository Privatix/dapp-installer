// +build linux darwin

package util

// #include <unistd.h>
import "C"

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"

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

// ExistingDappCtrlVersion returns existing dappctrl version.
func ExistingDappCtrlVersion(logger log.Logger) (string, bool) {
	logger.Warn(fmt.Sprintf("your OS %s is not supported now", runtime.GOOS))
	return "", false
}

// GrantAccess grants access to directory.
func GrantAccess(path, user string) error {
	cmd := exec.Command("chown", user, path)
	return cmd.Run()
}
