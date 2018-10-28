package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/privatix/dapp-openvpn/inst/pipeline"
	"github.com/privatix/dappctrl/util/log"

	"github.com/privatix/dapp-installer/dapp"
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
		flow = installFlow()
	case "update":
		logger.Info("update process")
		flow = updateFlow()
	case "remove":
		logger.Info("remove process")
		flow = removeFlow()
	case "help":
		fmt.Println(rootHelp)
		return
	default:
		processedRootFlags(version)
		return
	}

	d := dapp.NewConfig()

	if err := flow.Run(d, logger); err != nil {
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
		newOperator("install", install, nil),
		newOperator("remove temp", removeTemp, nil),
	}
}

func updateFlow() pipeline.Flow {
	return pipeline.Flow{
		newOperator("processed flags", processedUpdateFlags, nil),
		newOperator("validate", checkInstallation, nil),
		newOperator("init temp", initTemp, removeTemp),
		newOperator("update", install, nil),
		newOperator("remove temp", removeTemp, nil),
	}
}

func removeFlow() pipeline.Flow {
	return pipeline.Flow{
		newOperator("processed flags", processedRemoveFlags, nil),
		newOperator("validate", checkInstallation, nil),
		newOperator("stop services", stopServices, nil),
		newOperator("remove services", removeServices, nil),
		newOperator("remove dapp", removeDapp, nil),
	}
}
