package tor

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/privatix/dapp-installer/util"
)

// ServiceName returns tor daemon name.
func (t Tor) ServiceName() string {
	return "tor_" + util.Hash(t.RootPath)
}

// OnionName returns tor onion name.
func (t Tor) OnionName() (string, error) {
	f := filepath.Join(t.RootPath, "var", "lib", t.HiddenServiceDir,
		"hostname")

	done := make(chan bool)
	go func() {
		for {
			if _, err := os.Stat(f); err == nil {
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
		done <- true
	}()

	select {
	case <-done:
		d, err := ioutil.ReadFile(f)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(d)), nil
	case <-time.After(util.TimeOutInSec(30)):
		return "", errors.New("failed to read hostname. timeout expired")
	}
}

func installService(daemon, path, descr string) error {
	return nil
}

func startService(daemon string) error {
	return nil
}

func stopService(daemon string) error {
	return nil
}

func removeService(daemon string) error {
	return nil
}
