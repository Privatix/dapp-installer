package command

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

func createUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update dapp core to host",
		Long:  "Run upgrade process of dapp core to localhost",

		Run: update,
	}
}

func update(ccmd *cobra.Command, args []string) {
	log.Println("Start update process")
	fmt.Println("I will be running upgrading process", args)
	log.Println("Finish update process")
}
