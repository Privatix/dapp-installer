// +build darwin linux

package unix

// Daemon has a parameters for unix-daemon.
type Daemon struct {
	ID        string
	UID       string
	GUID      string
	Name      string
	Command   string
	Args      []string
	AutoStart bool
}

// NewDaemon creates a default deamon configuration.
func NewDaemon(id string) *Daemon {
	return &Daemon{
		ID:        id,
		AutoStart: true,
	}
}
