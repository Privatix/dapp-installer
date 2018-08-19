package command

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

func createRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove",
		Short: "Remove dapp core to host",
		Long:  "Run deinstallation process of dapp core to localhost",

		Run: remove,
	}
}

func remove(ccmd *cobra.Command, args []string) {
	log.Println("Start remove process")
	fmt.Println("I will be running uninstallation process", args)
	log.Println("Finish remove process")
}
