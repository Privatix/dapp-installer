package tor

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"syscall"

	"github.com/privatix/dapp-installer/util"
)

// ServiceName returns tor daemon name.
func (inst *Installation) ServiceName() string {
	return "tor_" + util.Hash(inst.RootPath)
}

func installService(daemon, path, descr string, autostart bool) error {
	path = filepath.Join(path, "var", "lib", "tor")

	s, err := os.Stat(path)
	if err != nil {
		return err
	}

	u, err := user.LookupId(fmt.Sprint(s.Sys().(*syscall.Stat_t).Uid))
	if err != nil {
		return err
	}

	p := filepath.Join(path, "hidden_service")
	return util.ExecuteCommand("chown", "-R", u.Username, p)
}

func startService(daemon, uid string) error {
	return nil
}

func stopService(daemon, uid string) error {
	return nil
}

func removeService(daemon string) error {
	return nil
}
