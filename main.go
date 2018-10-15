// dapp-installer is a cli installer for dapp core.

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

func printVersion() {
	fmt.Printf("dapp-installer %s %s", Version, Commit)
}

func main() {
	logger, closer, err := util.CreateLogger()
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %s", err))
	}
	defer closer.Close()

	logger.Info("============================================================")
	logger.Info("begin program")
	command.Execute(logger, printVersion, os.Args[1:])
	logger.Info("end program")
	fmt.Println("\nPress the Enter Key to close the console screen!")
	fmt.Scanln()
}
