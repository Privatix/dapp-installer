// dapp-installer is a cli installer for dapp core.
//
// For more specific usage information, refer to the help doc (dapp-installer -h):
//
//  dapp-installer - installer for dapp core
//
//  Usage:
//    dapp-installer [command] [flags]
//
//  Available Commands:
//    install     Install dapp core
//    update      Update dapp core
//    remove      Remove dapp core from host
//    info        Display info
//
//  Flags:
//    -h, --help      			 help for dapp-installer
//    -v, --version              Display the current version of this CLI
//
//  Use "dapp-installer [command] --help" for more information about a command.
//
package main

import (
	"fmt"

	"github.com/privatix/dapp-installer/command"
)

func main() {
	//first implementation for win platform
	// if platform.Version() != "windows" {
	// 	panic("Software install only to Windows platform")
	// }

	if err := command.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		return
	}

}
