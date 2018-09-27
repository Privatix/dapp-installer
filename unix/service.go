// +build linux darwin

package unix

import "errors"

// Service has a parameters for windows service.
type Service struct {
	ID          string
	GUID        string
	Name        string
	Description string
	Command     string
	Args        []string
	AutoStart   bool
}

// Uninstall uninstalls windows service.
func (s *Service) Uninstall() {
}

// Start starts windows service.
func (s *Service) Start() error {
	return errors.New("not implemented")
}

// Stop stops windows service.
func (s *Service) Stop() error {
	return errors.New("not implemented")
}

// Remove removes windows service.
func (s *Service) Remove() error {
	return errors.New("not implemented")
}
