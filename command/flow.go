// +build darwin windows

package command

import (
	"github.com/privatix/dapp-installer/pipeline"
)

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

func updateProductsFlow() pipeline.Flow {
	return pipeline.Flow{
		newOperator("processed flags", processedUpdateProductFlags, nil),
		newOperator("validate", checkInstallation, nil),
		newOperator("update products", updateProducts, nil),
		newOperator("start products", startProducts, nil),
	}
}

func removeProductsFlow() pipeline.Flow {
	return pipeline.Flow{
		newOperator("processed flags", processedRemoveProductFlags, nil),
		newOperator("validate", checkInstallation, nil),
		newOperator("remove products", removeProducts, nil),
	}
}
