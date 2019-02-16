package command

import (
	"runtime"

	"github.com/privatix/dapp-installer/pipeline"
)

func installFlow() pipeline.Flow {
	if runtime.GOOS == "linux" {
		return installLinuxFlow()
	}
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
	if runtime.GOOS == "linux" {
		return updateLinuxFlow()
	}
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
	if runtime.GOOS == "linux" {
		return removeLinuxFlow()
	}
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
	if runtime.GOOS == "linux" {
		return installLinuxProductsFlow()
	}
	return pipeline.Flow{
		newOperator("processed flags", processedInstallProductFlags, nil),
		newOperator("validate", checkInstallation, nil),
		newOperator("install products", installProducts, removeProducts),
	}
}

func updateProductsFlow() pipeline.Flow {
	if runtime.GOOS == "linux" {
		return updateLinuxProductsFlow()
	}
	return pipeline.Flow{
		newOperator("processed flags", processedUpdateProductFlags, nil),
		newOperator("validate", checkInstallation, nil),
		newOperator("update products", updateProducts, nil),
		newOperator("start products", startProducts, nil),
	}
}

func removeProductsFlow() pipeline.Flow {
	if runtime.GOOS == "linux" {
		return removeLinuxProductsFlow()
	}
	return pipeline.Flow{
		newOperator("processed flags", processedRemoveProductFlags, nil),
		newOperator("validate", checkInstallation, nil),
		newOperator("remove products", removeProducts, nil),
	}
}

func installLinuxFlow() pipeline.Flow {
	return pipeline.Flow{
		newOperator("processed flags", processedInstallFlags, nil),
		newOperator("validate", validateToInstall, nil),
		newOperator("prepare", prepare, nil),
		newOperator("init temp", initTemp, removeTemp),
		newOperator("extract", extract, removeDapp),
		newOperator("configure tor", installTor, nil),
		newOperator("configure dapp", configureDapp, nil),
		newOperator("install", installContainer, removeContainer),
		newOperator("start", startContainer, stopContainer),
		newOperator("create database", createDatabase, nil),
		newOperator("install products", installProducts, removeProducts),
		newOperator("finalize", finalize, nil),
		newOperator("remove temp", removeTemp, nil),
	}
}

func updateLinuxFlow() pipeline.Flow {
	return pipeline.Flow{
		newOperator("processed flags", processedUpdateFlags, nil),
		newOperator("validate", checkContainer, nil),
		newOperator("init temp", initTemp, removeTemp),
		newOperator("stop", stopContainer, startContainer),
		newOperator("update", updateContainer, restoreContainer),
		//newOperator("disable daemons", disableDaemons, nil),
		newOperator("start", startContainer, stopContainer),
		newOperator("update database", updateDatabase, nil),
		//newOperator("enable daemons", enableDaemons, nil),
		newOperator("finalize", finalize, nil),
		newOperator("remove temp", removeTemp, nil),
		newOperator("remove backup", removeBackup, nil),
	}
}

func removeLinuxFlow() pipeline.Flow {
	return pipeline.Flow{
		newOperator("processed flags", processedRemoveFlags, nil),
		newOperator("validate", checkContainer, nil),
		newOperator("stop", stopContainer, nil),
		newOperator("remove", removeContainer, nil),
		newOperator("remove dapp", removeDapp, nil),
	}
}

func installLinuxProductsFlow() pipeline.Flow {
	return pipeline.Flow{
		newOperator("processed flags", processedInstallProductFlags, nil),
		newOperator("validate", checkContainer, nil),
		newOperator("install products", installProducts, removeProducts),
		newOperator("finalize", finalize, nil),
	}
}

func updateLinuxProductsFlow() pipeline.Flow {
	return pipeline.Flow{
		newOperator("processed flags", processedUpdateProductFlags, nil),
		newOperator("validate", checkContainer, nil),
		newOperator("update products", updateProducts, nil),
		newOperator("start products", startProducts, nil),
		newOperator("finalize", finalize, nil),
	}
}

func removeLinuxProductsFlow() pipeline.Flow {
	return pipeline.Flow{
		newOperator("processed flags", processedRemoveProductFlags, nil),
		newOperator("validate", checkContainer, nil),
		newOperator("remove products", removeProducts, nil),
		newOperator("finalize", finalize, nil),
	}
}
