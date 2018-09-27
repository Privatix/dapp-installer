// +build linux darwin

package util

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/privatix/dappctrl/util/log"
)

// CheckSystemPrerequisites does checked system to prerequisites.
func CheckSystemPrerequisites(volume string, logger log.Logger) bool {
	logger.Warn(fmt.Sprintf("your OS %s is not supported now", runtime.GOOS))
	return false
}

// ExistingDappCtrlVersion returns existing dappctrl version.
func ExistingDappCtrlVersion(logger log.Logger) (string, bool) {
	logger.Warn(fmt.Sprintf("your OS %s is not supported now", runtime.GOOS))
	return "", false
}

// GrantAccess grants access to directory.
func GrantAccess(path, user string) error {
	return errors.New("not implemented")
}
