package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	removeCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove dapp core to host",
		Long:  "Run deinstallation process of dapp core to localhost",

		Run: remove,
	}
)

func remove(ccmd *cobra.Command, args []string) {
	fmt.Println("I will be run deinstall process", args)
}
