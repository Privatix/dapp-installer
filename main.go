// dapp-installer is a cli installer for dapp core.

//go:generate goversioninfo
package main

import (
	"fmt"
	"os"

	"github.com/privatix/dapp-installer/command"
	"github.com/privatix/dapp-installer/util"
)

// Values for versioning.
var (
	Commit  string
	Version string
)

const logFile = "dapp-installer.log"

func printVersion() {
	fmt.Printf("dapp-installer %s %s", Version, Commit)
}

func main() {
	file, logger, err := util.CreateLogger(logFile)
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %s", err))
	}

	defer file.Close()

	logger.Info("============================================================")
	logger.Info("begin program")
	command.Execute(logger, printVersion, os.Args[1:])
	logger.Info("end program")
	fmt.Println("Press the Enter Key to close the console screen!")
	fmt.Scanln()
}
