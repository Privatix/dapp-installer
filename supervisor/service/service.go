package service

import (
	"context"
	"time"

	"github.com/takama/daemon"
)

var actionTimeout = 15 * time.Second

const name = "Privatix_Dapp_Supervisor"

func supervisorService() (daemon.Daemon, error) {
	return daemon.New(name, "")
}

// Install install supervisor daemon to system.
func Install(args []string) error {
	args = append([]string{"runserver"}, args...)

	if service, err := supervisorService(); err != nil {
		return err
	} else if _, err := service.Install(args...); err != nil && err != daemon.ErrAlreadyInstalled { // Fault tolerance. Ignore if service is already installed.
		return err
	}
	return nil
}

// Start start supervisor daemon.
func Start(port int) error {
	service, err := supervisorService()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), actionTimeout)
	defer cancel()
	// Make sure daemon is started.
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, err = service.Start()
			if err != nil || (err == daemon.ErrAlreadyRunning || err == daemon.ErrNotInstalled) {
				return nil
			}
		}
	}
}

// Stop stops supervisor daemon.
func Stop() error {
	service, err := supervisorService()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), actionTimeout)
	defer cancel()
	// Make sure daemon is stopped.
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, err = service.Stop()
			if err != nil || (err == daemon.ErrAlreadyStopped || err == daemon.ErrNotInstalled) {
				return nil
			}
		}
	}
}

// Remove removes supervisor daemon from system.
func Remove() error {
	service, err := supervisorService()
	if err != nil {
		return err
	}
	// Make sure daemon is removed.
	for _, err := service.Remove(); err != nil && err != daemon.ErrNotInstalled; {
		return err
	}

	return nil
}
