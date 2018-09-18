// +build linux darwin

package dbengine

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/privatix/dappctrl/util/log"
)

// Install installs a DB engine.
func (engine *DBEngine) Install(path string, logger log.Logger) error {
	logger.Warn(fmt.Sprintf("your OS %s is not supported now", runtime.GOOS))
	return errors.New("is not supported for this feature")
}
