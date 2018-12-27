// +build linux

package unix

import "errors"

// Install installs the daemon.
func (d *Daemon) Install() error {
	return errors.New("not implemented")
}

// Start starts the daemon.
func (d *Daemon) Start() error {
	return errors.New("not implemented")
}

// Stop stops the daemon.
func (d *Daemon) Stop() error {
	return errors.New("not implemented")
}

// Remove removes the daemon.
func (d *Daemon) Remove() error {
	return errors.New("not implemented")
}

// IsStopped returns the daemon stopped status.
func (d *Daemon) IsStopped() bool {
	return true
}
