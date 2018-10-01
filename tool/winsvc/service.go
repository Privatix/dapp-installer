// +build windows

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/privatix/dappctrl/util/log"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

type service struct {
	pid     *os.Process
	log     *eventlog.Log
	running bool
}

func (s *service) Execute(args []string, r <-chan svc.ChangeRequest,
	changes chan<- svc.Status) (ssec bool, errno uint32) {
	const acceptedCmds = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: acceptedCmds}
	s.running = true
loop:
	for c := range r {
		switch c.Cmd {
		case svc.Interrogate:
			changes <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			break loop
		default:
			s.log.Error(1, fmt.Sprintf("unexpected control request: %d",
				c.Cmd))
		}
	}

	// kill child process
	if s.pid != nil {
		//s.log.Info(1, fmt.Sprintf("kill process %v", s.pid))
		if err := s.pid.Kill(); err != nil {
			s.log.Error(1, fmt.Sprintf(
				"failed to stop child process: %v", err))
		}
	}
	//s.log.Info(1, fmt.Sprintf("stoped %v", s))
	changes <- svc.Status{State: svc.StopPending}
	return
}

func installService(svcConf *serviceConfig) error {
	exepath, err := exePath()
	if err != nil {
		return err
	}

	ok := existService(svcConf.ID)
	if ok {
		return fmt.Errorf("service %s already exists", svcConf.ID)
	}

	return registerService(svcConf, exepath)
}

func runService(name string, s *service) {
	log, err := eventlog.Open(name)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer log.Close()

	log.Info(1, fmt.Sprintf("starting %s service", name))
	s.log = log
	if err := svc.Run(name, s); err != nil {
		log.Error(1, fmt.Sprintf("%s service failed %v", name, err))
		return
	}
	log.Info(1, fmt.Sprintf("%s service stopped", name))
}

func removeService(name string) error {
	m, s, err := getService(name)
	if err != nil {
		return fmt.Errorf("service %s is not installed: %v", name, err)
	}
	defer m.Disconnect()
	defer s.Close()

	err = s.Delete()
	if err != nil {
		return err
	}
	err = eventlog.Remove(name)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %s", err)
	}
	return nil
}

func registerService(svcConf *serviceConfig, exepath string) error {
	var startType uint32 = mgr.StartManual
	if svcConf.AutoStart {
		startType = mgr.StartAutomatic
	}

	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.CreateService(svcConf.ID, exepath,
		mgr.Config{
			DisplayName: svcConf.Name,
			StartType:   startType,
			Description: svcConf.Description,
		})
	if err != nil {
		return err
	}
	defer s.Close()
	err = eventlog.InstallAsEventCreate(svcConf.ID,
		eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("SetupEventLogSource() failed: %s", err)
	}
	return nil
}

func getService(name string) (*mgr.Mgr, *mgr.Service, error) {
	m, err := mgr.Connect()
	if err != nil {
		return nil, nil, err
	}
	s, err := m.OpenService(name)
	if err != nil {
		m.Disconnect()
		return nil, nil, err
	}
	return m, s, nil
}

func existService(name string) bool {
	m, err := mgr.Connect()
	if err != nil {
		return false
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return false
	}
	defer s.Close()
	return true
}

func startService(name string) error {
	m, s, err := getService(name)
	if err != nil {
		return fmt.Errorf("could not access to service: %v", err)
	}
	defer m.Disconnect()
	defer s.Close()
	if err := s.Start(); err != nil {
		return fmt.Errorf("could not start service: %v", err)
	}
	return nil
}

func stopService(name string) error {
	m, s, err := getService(name)
	if err != nil {
		return fmt.Errorf("could not access to service: %v", err)
	}
	defer m.Disconnect()
	defer s.Close()
	status, err := s.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf("could not stop service: %v", err)
	}
	timeout := time.Now().Add(10 * time.Second)
	for status.State != svc.Stopped {
		if timeout.Before(time.Now()) {
			return fmt.Errorf("timeout waiting for service to go to state=%d",
				svc.Stopped)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return fmt.Errorf("could not retrieve service status: %v", err)
		}
	}
	return nil
}

func exePath() (string, error) {
	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = fmt.Errorf("%s is directory", p)
	}
	if filepath.Ext(p) == "" {
		p += ".exe"
		fi, err := os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			err = fmt.Errorf("%s is directory", p)
		}
	}
	return "", err
}

func runServiceCommand(conf *serviceConfig, logger log.Logger, file *os.File) {
	cmd := exec.Command(conf.Command, conf.Args...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create stderrpipe: %v", err))
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create stdoutpipe: %v", err))
		return
	}

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	logger.Info("starting child process")
	if err := cmd.Start(); err != nil {
		logger.Error(fmt.Sprintf("failed to start child process: %v", err))
		return
	}
	conf.service.pid = cmd.Process

	go io.Copy(writer, stdout)
	go io.Copy(writer, stderr)

	if err := cmd.Wait(); err != nil {
		logger.Error(fmt.Sprintf("failed to wait child process: %v", err))
		return
	}
}
