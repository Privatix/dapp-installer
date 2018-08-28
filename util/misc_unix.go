// +build linux darwin

package util

import (
	"fmt"
	"runtime"

	"github.com/Privatix/dappctrl/util/log"
)

// CheckSystemPrerequisites does checked system to prerequisites
func CheckSystemPrerequisites(log log.Logger) bool {
	log.Warn(fmt.Sprintf("your OS %s is not supported now", runtime.GOOS))
	return false
}
