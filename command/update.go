package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update dapp core to host",
		Long:  "Run upgrade process of dapp core to localhost",

		Run: update,
	}
)

func update(ccmd *cobra.Command, args []string) {
	fmt.Println("I will be run upgrade process", args)
}
