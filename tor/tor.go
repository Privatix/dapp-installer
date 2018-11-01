package tor

// Tor has a tor configuration.
type Tor struct {
	Service  string
	Settings string
}

// NewTor creates a default Tor configuration.
func NewTor() *Tor {
	return &Tor{}
}

func (t *Tor) configurate(path string) error {

	return nil
}

// Install installs tor process.
func (t *Tor) Install(path string) error {
	// todo configure
	return installService(serviceName(path), path)
}

// Start starts tor process.
func (t *Tor) Start(path string) error {
	return startService(serviceName(path))
}

// Stop stops tor process.
func (t *Tor) Stop(path string) error {
	return stopService(serviceName(path))
}

// Remove removes tor process.
func (t *Tor) Remove(path string) error {
	return removeService(serviceName(path))
}
