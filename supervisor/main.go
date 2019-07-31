package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/privatix/dapp-installer/dapp"
	"github.com/privatix/dapp-installer/supervisor/server"
	"github.com/privatix/dapp-installer/supervisor/service"
	"github.com/privatix/dappctrl/util/log"
)

func createLogger(dir string) (log.Logger, io.Closer) {
	failIfErr := func(err error) {
		if err != nil {
			panic(fmt.Sprintf("failed to create logger: %v", err))
		}
	}

	logConfig := &log.FileConfig{
		WriterConfig: log.NewWriterConfig(),
		Filename:     filepath.Join(dir, "dapp-supervisor-%Y-%m-%d.log"),
		FileMode:     0644,
	}

	flog, closer, err := log.NewFileLogger(logConfig)
	failIfErr(err)

	logger := flog

	for _, value := range os.Args[1:] {
		if value == "-verbose" || value == "--verbose" {
			elog, err := log.NewStderrLogger(log.NewWriterConfig())
			failIfErr(err)
			logger = log.NewMultiLogger(elog, flog)
			break
		}
	}

	return logger, closer
}

func main() {
	if len(os.Args) == 1 {
		panic("missing required argument")
	}

	dir, _ := filepath.Split(os.Args[0])
	logger, closer := createLogger(dir)
	defer closer.Close()

	logger.Info(fmt.Sprint("exec with args: ", os.Args[1:]))

	arg := os.Args[1]
	switch arg {
	case "run-service":
		// Non blocking service start. Some systems (windows) expect a service
		// to answer to start control in a timely fashion, otherwise it is
		// considered irresponsive and killed.
		args := append([]string{"runserver"}, os.Args[2:]...)
		if err := service.RunServer(logger, os.Args[0], args); err != nil {
			logger.Fatal(err.Error())
		}
	case "runserver":
		// Expected os.Args = dapp-supervisor runserver <role> <path> <tor.rootpath>
		d := dapp.NewDapp()
		port := os.Args[2]
		d.Role = os.Args[3]
		d.Path = os.Args[4]
		d.Controller.Service.ID = d.ControllerHash()
		d.Controller.Service.Name = d.Controller.Service.ID
		d.Tor.RootPath = os.Args[5]
		d.Controller.Service.GUID = filepath.Join(d.Path, "dappctrl", d.Controller.Service.ID)
		var installUID string // Passed only for darwin.
		if runtime.GOOS == "darwin" {
			installUID = os.Args[6]
		}
		d.Controller.Service.SetUID(installUID)
		logger.Fatal(server.ListenAndServe(logger, "127.0.0.1:"+port, d, installUID).Error())
	case "install":
		// Make sure port is valid integer before installing.
		if _, err := strconv.Atoi(os.Args[2]); err != nil {
			logger.Fatal(fmt.Sprintf("could not parse port number: %v", err))
		}
		args := append([]string{"run-service"}, os.Args[2:]...)
		if err := service.Install(args); err != nil {
			logger.Fatal(err.Error())
		}
		if err := service.Start(); err != nil {
			// Could not start, cleaning up.
			if err := service.Remove(); err != nil {
				logger.Error(err.Error())
			}
			logger.Fatal(err.Error())
		}
	case "remove":
		if err := service.Stop(); err != nil {
			logger.Fatal(err.Error())
		}
		if err := service.Remove(); err != nil {
			logger.Fatal(err.Error())
		}
	}
}
