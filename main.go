// dapp-installer is a cli installer for dapp core.

//go:generate goversioninfo
package main

import (
	"fmt"
	"os"

	"github.com/Privatix/dapp-installer/command"
	"github.com/Privatix/dappctrl/util/log"
)

// Values for versioning.
var (
	Commit  string
	Version string
)

const logFile = "dapp-installer.log"

type config struct {
	DB string
}

func newConfig() *config {
	return &config{
		DB: "db config",
	}
}

func printVersion() {
	fmt.Printf("dapp-installer %s %s", Version, Commit)
}

func createLogger() (*os.File, log.Logger, error) {
	file, err := os.OpenFile(
		logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, nil, err
	}

	logger, err := log.NewFileLogger(log.NewFileConfig(), file)
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	return file, logger, nil
}

func main() {
	file, logger, err := createLogger()
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %s", err))
	}

	defer file.Close()

	logger.Info("begin program")
	command.Execute(logger, printVersion, os.Args[1:])
	logger.Info("end program")
}
