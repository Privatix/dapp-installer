// +build linux darwin

package util

import (
	"fmt"
	"runtime"

	"github.com/privatix/dappctrl/util/log"
)

// CheckSystemPrerequisites does checked system to prerequisites.
func CheckSystemPrerequisites(volume string, logger log.Logger) bool {
	logger.Warn(fmt.Sprintf("your OS %s is not supported now", runtime.GOOS))
	return false
}

// ExistingDBEnginePort returns existing db engine port number.
func ExistingDBEnginePort(logger log.Logger) (int, bool) {
	logger.Warn(fmt.Sprintf("your OS %s is not supported now", runtime.GOOS))
	return 0, false
}

// ExistingDappCtrlVersion returns existing dappctrl version.
func ExistingDappCtrlVersion(logger log.Logger) (string, bool) {
	logger.Warn(fmt.Sprintf("your OS %s is not supported now", runtime.GOOS))
	return "", false
}
