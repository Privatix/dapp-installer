// dapp-installer is a cli installer for dapp core.

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/privatix/dappctrl/util/log"
	"github.com/privatix/dappctrl/version"

	"github.com/privatix/dapp-installer/command"
)

// Values for versioning.
var (
	Commit  string
	Version string
)

func printVersion() {
	version.Print(true, Commit, Version)
}

func createLogger() (log.Logger, io.Closer, error) {
	elog, err := log.NewStderrLogger(log.NewWriterConfig())
	if err != nil {
		return nil, nil, err
	}

	logConfig := &log.FileConfig{
		WriterConfig: log.NewWriterConfig(),
		Filename:     "dapp-installer-%Y-%m-%d.log",
		FileMode:     0644,
	}

	flog, closer, err := log.NewFileLogger(logConfig)
	if err != nil {
		return nil, nil, err
	}

	logger := log.NewMultiLogger(elog, flog)

	return logger, closer, nil
}

func main() {
	logger, closer, err := createLogger()
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %s", err))
	}
	defer closer.Close()

	command.Execute(logger, printVersion, os.Args[1:])

	fmt.Println("\nPress the Enter Key to close the console screen!")
	fmt.Scanln()
}
