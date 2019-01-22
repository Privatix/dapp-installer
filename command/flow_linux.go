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
		newOperator("configure", configure, nil),
		newOperator("install", installContainer, removeContainer),
		newOperator("start", startContainer, stopContainer),
		newOperator("finalize", finalize, nil),
		newOperator("remove temp", removeTemp, nil),
	}
}

func updateFlow() pipeline.Flow {
	return nil
}

func removeFlow() pipeline.Flow {
	return nil
}

func statusFlow() pipeline.Flow {
	return nil
}

func installProductsFlow() pipeline.Flow {
	return nil
}

func updateProductsFlow() pipeline.Flow {
	return nil
}

func removeProductsFlow() pipeline.Flow {
	return nil
}
