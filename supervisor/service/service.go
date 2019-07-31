package service

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/takama/daemon"

	"github.com/privatix/dappctrl/util/log"
)

var actionTimeout = 15 * time.Second

const name = "Privatix_Dapp_Supervisor"

func supervisorService() (daemon.Daemon, error) {
	return daemon.New(name, "")
}

// Install install supervisor daemon to system.
func Install(args []string) error {
	if service, err := supervisorService(); err != nil {
		return err
	} else if _, err := service.Install(args...); err != nil && err != daemon.ErrAlreadyInstalled { // Fault tolerance. Ignore if service is already installed.
		return err
	}
	return nil
}

// Start start supervisor daemon.
func Start() error {
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

type executable struct {
	logger  log.Logger
	exec    string
	args    []string
	process *os.Process
	mu      sync.Mutex
}

// Run blocking server run.
func (e *executable) Run() {
	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
	defer close(interrupt)

	e.Start()

	for {
		select {
		case <-interrupt:
			e.Stop()
		}
	}
}

// Start non-blocking run.
func (e *executable) Start() {
	go func() {
		if err := e.run(); err != nil {
			e.logger.Error(err.Error())
			os.Exit(2)
		}
		os.Exit(0)
	}()
}

// Stop non-blocking stop.
func (e *executable) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.process != nil {
		if err := e.process.Kill(); err != nil {
			e.logger.Error(err.Error())
			e.process = nil
		}
	} else {
		e.logger.Warn("couldn't stop: nothing to stop")
	}
}

func (e *executable) run() error {
	e.mu.Lock()
	e.mu.Unlock()
	cmd := exec.Command(e.exec, e.args...)

	if err := cmd.Start(); err != nil {
		e.logger.Error(err.Error())
		return err
	}

	e.process = cmd.Process

	return cmd.Wait()
}

// RunServer starts a server.
func RunServer(logger log.Logger, exec string, args []string) error {
	service, err := supervisorService()
	if err != nil {
		return err
	}

	_, err = service.Run(&executable{
		logger: logger,
		exec:   exec,
		args:   args,
	})

	return err
}
