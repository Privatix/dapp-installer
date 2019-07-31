// +build linux

package unix

// SetUID sets uid.
func (d *Daemon) SetUID(uid string) {
	d.UID = uid
}

// Install installs the daemon.
func (d *Daemon) Install() error {
	return nil
}

// Start starts the daemon.
func (d *Daemon) Start() error {
	return nil
}

// Stop stops the daemon.
func (d *Daemon) Stop() error {
	return nil
}

// Remove removes the daemon.
func (d *Daemon) Remove() error {
	return nil
}

// IsStopped returns the daemon stopped status.
func (d *Daemon) IsStopped() bool {
	return true
}
