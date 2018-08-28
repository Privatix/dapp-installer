// +build linux darwin

package util

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/Privatix/dappctrl/util/log"
)

// CheckSystemPrerequisites does checked system to prerequisites.
func CheckSystemPrerequisites(log log.Logger) bool {
	log.Warn(fmt.Sprintf("your OS %s is not supported now", runtime.GOOS))
	return false
}

// DBEngineExists is checking to install DB engine.
func DBEngineExists(log log.Logger) bool {
	log.Warn(fmt.Sprintf("your OS %s is not supported now", runtime.GOOS))
	return false
}

// InstallDBEngine is installing DB engine.
func InstallDBEngine(dbConf *DBEngine, log log.Logger) error {
	log.Warn(fmt.Sprintf("your OS %s is not supported now", runtime.GOOS))
	return errors.New("is not supported for this feature")
}
