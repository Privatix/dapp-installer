package service

import (
	"context"
	"os"
	"runtime"
	"time"

	"github.com/privatix/dapp-installer/unix"
	"github.com/takama/daemon"
)

var actionTimeout = 15 * time.Second

const name = "Privatix Dapp-Superviser"

func supervisorService() (daemon.Daemon, error) {
	return daemon.New(name, "")
}

// Install install supervisor daemon to system.
func Install(args []string) error {
	args = append([]string{"runserver"}, args...)

	// github.com/takama/daemon install services with root privilege. For darwin need as current users.
	if runtime.GOOS == "darwin" {
		d := unix.NewDaemon(name)
		d.Command = os.Args[0]
		d.Args = args

		return d.Install()
	}

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
			if runtime.GOOS == "darwin" {
				if err := unix.NewDaemon(name).Start(); err == nil {
					return nil
				}
			} else {
				_, err = service.Start()
				if err == daemon.ErrAlreadyRunning || err == daemon.ErrNotInstalled {
					return nil
				}
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
			if runtime.GOOS == "darwin" {
				if err = unix.NewDaemon(name).Stop(); err == nil {
					return nil
				}
			}
			_, err = service.Stop()
			if err == daemon.ErrAlreadyStopped || err == daemon.ErrNotInstalled {
				return nil
			}
		}
	}
}

// Remove removes superviser daemon from system.
func Remove() error {
	if runtime.GOOS == "darwin" {
		return unix.NewDaemon(name).Remove()
	}
	service, err := supervisorService()
	if err != nil {
		return err
	}
	// Make sure daemon is removed.
	for _, err := service.Remove(); err != nil && err != daemon.ErrNotInstalled; {
	}

	return nil
}
