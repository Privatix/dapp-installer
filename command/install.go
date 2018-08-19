package command

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

func createInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install dapp core to host",
		Long:  "Run installation process of dapp core to localhost",

		Run: install,
	}
}

func install(ccmd *cobra.Command, args []string) {
	log.Println("Start install process")
	fmt.Println("I will be runnning installation process", args)
	log.Println("Finish install process")
}
