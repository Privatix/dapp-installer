package command

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/privatix/dappctrl/util/log"

	"github.com/privatix/dapp-installer/dapp"
	"github.com/privatix/dapp-installer/pipeline"
)

// Execute executes a CLI command.
func Execute(logger log.Logger, version func(), args []string) {
	if len(args) == 0 {
		args = append(args, "help")
	}

	var flow pipeline.Flow

	switch strings.ToLower(args[0]) {
	case "install":
		logger.Info("install process")
		logger = logger.Add("action", "install")
		flow = installFlow()
	case "install-products":
		logger.Info("install products process")
		logger = logger.Add("action", "install products")
		flow = installProductsFlow()
	case "update":
		logger.Info("update process")
		logger = logger.Add("action", "update")
		flow = updateFlow()
	case "update-products":
		logger.Info("update products process")
		logger = logger.Add("action", "update products")
		flow = updateProductsFlow()
	case "remove":
		logger.Info("remove process")
		logger = logger.Add("action", "remove")
		flow = removeFlow()
	case "remove-products":
		logger.Info("remove products process")
		logger = logger.Add("action", "remove products")
		flow = removeProductsFlow()
	case "status":
		logger.Info("status process")
		logger = logger.Add("action", "status")
		flow = statusFlow()
	case "help":
		fmt.Println(rootHelp)
		return
	default:
		processedRootFlags(version)
		return
	}

	d := dapp.NewDapp()

	if err := flow.Run(d, logger); err != nil {
		object, _ := json.Marshal(d)
		logger = logger.Add("object", string(object))
		logger.Error(fmt.Sprintf("%v", err))
		if !d.Verbose {
			fmt.Println("error:", err)
			fmt.Println("object:", string(object))
		}
		os.Exit(2)
	}

	logger.Info("command was successfully executed")
}
