package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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

	arg := os.Args[1]
	switch arg {
	case "runserver":
		// Expected os.Args = dapp-supervisor runserver <role> <path> <tor.rootpath>
		d := dapp.NewDapp()
		port := os.Args[2]
		d.Role = os.Args[3]
		d.Path = os.Args[4]
		d.Controller.Service.ID = d.ControllerHash()
		d.Tor.RootPath = os.Args[5]
		logger.Fatal(server.ListenAndServe(logger, "127.0.0.1:"+port, d).Error())
	case "install":
		port := os.Args[2]
		// Make sure port is valid integer before installing.
		portInt, err := strconv.Atoi(port)
		if err != nil {
			logger.Fatal(fmt.Sprintf("could not parse port number: %v", err))
		}
		if err := service.Install(os.Args[2:]); err != nil {
			logger.Fatal(err.Error())
		}
		if err := service.Start(portInt); err != nil {
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
