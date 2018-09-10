// +build linux darwin

package util

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/Privatix/dappctrl/util/log"
)

// CheckSystemPrerequisites does checked system to prerequisites.
func CheckSystemPrerequisites(logger log.Logger) bool {
	logger.Warn(fmt.Sprintf("your OS %s is not supported now", runtime.GOOS))
	return false
}

// ExistingDBEnginePort returns existing db engine port number.
func ExistingDBEnginePort(logger log.Logger) (int, bool) {
	logger.Warn(fmt.Sprintf("your OS %s is not supported now", runtime.GOOS))
	return 0, false
}

// InstallDBEngine is installing DB engine.
func InstallDBEngine(dbConf *DBEngine, logger log.Logger) error {
	logger.Warn(fmt.Sprintf("your OS %s is not supported now", runtime.GOOS))
	return errors.New("is not supported for this feature")
}

// ExistingDappCtrlVersion returns existing dappctrl version.
func ExistingDappCtrlVersion(logger log.Logger) (string, bool) {
	logger.Warn(fmt.Sprintf("your OS %s is not supported now", runtime.GOOS))
	return "", false
}
