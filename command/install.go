package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	installCmd = &cobra.Command{
		Use:   "install",
		Short: "Install dapp core to host",
		Long:  "Run installation process of dapp core to localhost",

		Run: install,
	}
)

func install(ccmd *cobra.Command, args []string) {
	fmt.Println("I will be run install process", args)
}
