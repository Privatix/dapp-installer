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
		os.Exit(2)
	}

	logger.Info("command was successfully executed")
}

func installFlow() pipeline.Flow {
	return pipeline.Flow{
		newOperator("processed flags", processedInstallFlags, nil),
		newOperator("validate", validateToInstall, nil),
		newOperator("init temp", initTemp, removeTemp),
		newOperator("extract", extract, removeDapp),
		newOperator("install tor", installTor, removeTor),
		newOperator("start tor", startTor, stopTor),
		newOperator("install dbengine", installDBEngine, removeDBEngine),
		newOperator("install", install, remove),
		newOperator("install products", installProducts, removeProducts),
		newOperator("write version", writeVersion, nil),
		newOperator("write env", writeEnvironmentVariable, nil),
		newOperator("remove temp", removeTemp, nil),
	}
}

func updateFlow() pipeline.Flow {
	return pipeline.Flow{
		newOperator("processed flags", processedUpdateFlags, nil),
		newOperator("validate", checkInstallation, nil),
		newOperator("init temp", initTemp, removeTemp),
		newOperator("stop tor", stopTor, startTor),
		newOperator("stop services", stopServices, startServices),
		newOperator("update", update, startProducts),
		newOperator("write version", writeVersion, nil),
		newOperator("write env", writeEnvironmentVariable, nil),
		newOperator("start tor", startTor, nil),
		newOperator("start products", startProducts, nil),
		newOperator("remove temp", removeTemp, nil),
	}
}

func removeFlow() pipeline.Flow {
	return pipeline.Flow{
		newOperator("processed flags", processedRemoveFlags, nil),
		newOperator("validate", checkInstallation, nil),
		newOperator("stop services", stopServices, nil),
		newOperator("stop tor", stopTor, nil),
		newOperator("remove products", removeProducts, nil),
		newOperator("remove services", removeServices, nil),
		newOperator("remove tor", removeTor, nil),
		newOperator("remove dapp", removeDapp, nil),
	}
}

func statusFlow() pipeline.Flow {
	return pipeline.Flow{
		newOperator("processed flags", processedStatusFlags, nil),
		newOperator("print status", printStatus, nil),
	}
}

func installProductsFlow() pipeline.Flow {
	return pipeline.Flow{
		newOperator("processed flags", processedInstallProductFlags, nil),
		newOperator("validate", checkInstallation, nil),
		newOperator("install products", installProducts, removeProducts),
	}
}

func removeProductsFlow() pipeline.Flow {
	return pipeline.Flow{
		newOperator("processed flags", processedRemoveProductFlags, nil),
		newOperator("validate", checkInstallation, nil),
		newOperator("remove products", removeProducts, nil),
	}
}
