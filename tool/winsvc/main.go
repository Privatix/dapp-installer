// +build windows

// This is wrapper tool for create windows service.
// Windows services are configured by configuration files.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	dapputil "github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
	"golang.org/x/sys/windows/svc"
)

type serviceConfig struct {
	ID          string
	Name        string
	Description string
	Command     string
	Args        []string
	AutoStart   bool
	service     *service
}

func newServiceConfig() *serviceConfig {
	return &serviceConfig{
		ID:          "dappservice",
		Name:        "dappservice",
		Description: "dappservice description",
		AutoStart:   true,
		service:     &service{},
	}
}
func readFlags(conf *serviceConfig) {
	fileName := os.Args[0][0 : len(os.Args[0])-len(filepath.Ext(os.Args[0]))]
	configFilePath := fmt.Sprintf("%s.config.json", fileName)
	fconfig := flag.String("config", configFilePath, "Configuration file")

	flag.Parse()

	if err := dapputil.ReadJSONFile(*fconfig, &conf); err != nil {
		panic(fmt.Sprintf("failed to read configuration file: %s", err))
	}
}

func main() {
	conf := newServiceConfig()
	readFlags(conf)

	logConfig := &log.FileConfig{
		WriterConfig: log.NewWriterConfig(),
		Filename:     "dapp-winservice-%Y-%m-%d.log",
		FileMode:     0644,
	}

	logger, closer, err := log.NewFileLogger(logConfig)
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %s", err))
	}

	defer closer.Close()

	logger = logger.Add("dapp-winservice", conf.ID)
	logger.Info("begin")
	defer logger.Info("end")
	logger.Info(fmt.Sprintf("%v", conf))
	executeCommand(conf, logger, os.Args[1:])
}

func executeCommand(svcConf *serviceConfig, logger log.Logger, args []string) {
	isIntSess, err := svc.IsAnInteractiveSession()
	if err != nil {
		logger.Error(fmt.Sprintf(
			"failed to determinate running session: %v", err))
	}

	if !isIntSess {
		go runService(svcConf.ID, svcConf.service)
		for {
			time.Sleep(200 * time.Millisecond)
			if svcConf.service.running {
				break
			}
		}
		runServiceCommand(svcConf, logger)
		return
	}

	if len(args) == 0 {
		return
	}

	switch args[0] {
	case "install":
		logger.Info(fmt.Sprintf("installation %s service", svcConf.ID))
		err = installService(svcConf)
	case "remove":
		logger.Info(fmt.Sprintf("removing %s service", svcConf.ID))
		err = removeService(svcConf.ID)
	case "start":
		logger.Info(fmt.Sprintf("starting %s service", svcConf.ID))
		err = startService(svcConf.ID)
	case "stop":
		logger.Info(fmt.Sprintf("stoping %s service", svcConf.ID))
		err = stopService(svcConf.ID)
	default:
		return
	}

	if err != nil {
		logger.Error(err.Error())
	}
	os.Exit(0)
}
