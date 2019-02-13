package command

import (
	"github.com/privatix/dapp-installer/pipeline"
)

func installFlow() pipeline.Flow {
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

func updateFlow() pipeline.Flow {
	return nil
}

func removeFlow() pipeline.Flow {
	return pipeline.Flow{
		newOperator("processed flags", processedRemoveFlags, nil),
		newOperator("validate", checkContainer, nil),
		newOperator("stop", stopContainer, nil),
		newOperator("remove", removeContainer, nil),
		newOperator("remove dapp", removeDapp, nil),
	}
}

func statusFlow() pipeline.Flow {
	return nil
}

func installProductsFlow() pipeline.Flow {
	return pipeline.Flow{
		newOperator("processed flags", processedInstallProductFlags, nil),
		newOperator("validate", checkContainer, nil),
		newOperator("install products", installProducts, removeProducts),
		newOperator("finalize", finalize, nil),
	}
}

func updateProductsFlow() pipeline.Flow {
	return pipeline.Flow{
		newOperator("processed flags", processedUpdateProductFlags, nil),
		newOperator("validate", checkContainer, nil),
		newOperator("update products", updateProducts, nil),
		newOperator("start products", startProducts, nil),
		newOperator("finalize", finalize, nil),
	}
}

func removeProductsFlow() pipeline.Flow {
	return pipeline.Flow{
		newOperator("processed flags", processedRemoveProductFlags, nil),
		newOperator("validate", checkContainer, nil),
		newOperator("remove products", removeProducts, nil),
		newOperator("finalize", finalize, nil),
	}
}
