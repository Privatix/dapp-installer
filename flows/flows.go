package flows

import (
	"runtime"

	"github.com/privatix/dapp-installer/flow"
)

// Install is core installation flow.
func Install() flow.Flow {
	if runtime.GOOS == "linux" {
		return InstallLinux()
	}
	return flow.Flow{
		Name: "Installation",
		Steps: []flow.Step{
			newStep("process flags", processedInstallFlags, nil),
			newStep("validation", validateToInstall, nil),
			newStep("init temp", initTemp, removeTemp),
			newStep("extract and update version", extractAndUpdateVersion, removeDapp),
			newStep("install tor", installTor, removeTor),
			newStep("start tor", startTor, stopTor),
			newStep("install dbengine", installDBEngine, removeDBEngine),
			newStep("install", install, remove),
			newStep("install products", installProducts, removeProducts),
			newStep("install supervisor if client", installSupervisorIfClient, removeSupervisorIfClient),
			newStep("update sendremote setting", updateSendRemote, nil),
			newStep("write env", writeEnvironmentVariable, nil),
			newStep("remove temp", removeTemp, nil),
			newStep("stop tor if client", stopTorIfClient, startTorIfClient),
			newStep("stop products if client", stopProductsIfClient, startProductsIfClient),
			newStep("stop dappctrl and db engine if client", stopServicesIfClient, startServicesIfClient),
		},
	}
}

// Remove is remove flow.
func Remove() flow.Flow {
	if runtime.GOOS == "linux" {
		return RemoveLinux()
	}
	return flow.Flow{
		Name: "Remove",
		Steps: []flow.Step{
			newStep("process flags", processedRemoveFlags, stopServicesIfClient),
			newStep("check installation", checkInstallation, nil),
			newStep("stop services", stopServices, nil),
			newStep("stop tor", stopTor, nil),
			newStep("remove supervisor", removeSupervisorIfClient, installSupervisorIfClient),
			newStep("remove products", removeProducts, nil),
			newStep("remove services", removeServices, nil),
			newStep("remove tor", removeTor, nil),
			newStep("remove dapp", removeDapp, nil),
		},
	}
}

// Status is display status flow.
func Status() flow.Flow {
	return flow.Flow{
		Name: "Status",
		Steps: []flow.Step{
			newStep("process flags", processedStatusFlags, nil),
			newStep("print status", printStatus, nil),
		},
	}
}

// InstallProducts is products installatin flow.
func InstallProducts() flow.Flow {
	if runtime.GOOS == "linux" {
		return InstallLinuxProducts()
	}
	return flow.Flow{
		Name: "Products installation",
		Steps: []flow.Step{
			newStep("process flags", processedInstallProductFlags, nil),
			newStep("validate", checkInstallation, nil),
			newStep("install products", installProducts, removeProducts),
			newStep("stop products if client", stopProductsIfClient, startProductsIfClient),
			newStep("stop dappctrl and db engine if client", stopServicesIfClient, startServicesIfClient),
		},
	}
}

// RemoveProducts is products remove flow.
func RemoveProducts() flow.Flow {
	if runtime.GOOS == "linux" {
		return RemoveLinuxProducts()
	}
	return flow.Flow{
		Name: "Products remove",
		Steps: []flow.Step{
			newStep("process flags", processedRemoveProductFlags, nil),
			newStep("validate", checkInstallation, nil),
			newStep("remove products", removeProducts, nil),
		},
	}
}

// InstallLinux is core installation flow for linux.
func InstallLinux() flow.Flow {
	return flow.Flow{
		Name: "Installation linux",
		Steps: []flow.Step{
			newStep("process flags", processedInstallFlags, nil),
			newStep("validate", validateToInstall, nil),
			newStep("prepare", prepare, nil),
			newStep("init temp", initTemp, removeTemp),
			newStep("extract", extractAndUpdateVersion, removeDapp),
			newStep("configure tor", installTor, nil),
			newStep("configure dapp", configureDapp, nil),
			newStep("install container", installContainer, removeContainer),
			newStep("enable and start container", enableAndStartContainer, disableAndStopContainer),
			newStep("create database", createDatabase, nil),
			newStep("install products", installProducts, removeProducts),
			newStep("disable container if client", disablContainerIfClient, nil),
			newStep("install supervisor", installSupervisorIfClient, removeSupervisorIfClient),
			newStep("check container running", finalize, nil),
			newStep("stop container if client", stopContainerIfClient, startContainerIfClient),
			newStep("remove temp", removeTemp, nil),
		},
	}
}

// RemoveLinux is remove flow for linux.
func RemoveLinux() flow.Flow {
	return flow.Flow{
		Name: "removeLinux",
		Steps: []flow.Step{
			newStep("process flags", processedRemoveFlags, nil),
			newStep("start container if client", startContainerIfClient, stopContainerIfClient),
			newStep("validate", checkContainer, nil),
			newStep("stop", stopContainer, nil),
			newStep("remove supervisor", removeSupervisorIfClient, installSupervisorIfClient),
			newStep("remove", removeContainer, nil),
			newStep("remove dapp", removeDapp, nil),
		},
	}
}

// InstallLinuxProducts is products installation flow for linux.
func InstallLinuxProducts() flow.Flow {
	return flow.Flow{
		Name: "Products installation (linux)",
		Steps: []flow.Step{
			newStep("process flags", processedInstallProductFlags, nil),
			newStep("start container if client", startContainerIfClient, stopContainerIfClient),
			newStep("validate", checkContainer, nil),
			newStep("install products", installProducts, removeProducts),
			newStep("finalize", finalize, nil),
			newStep("stop if client", stopContainerIfClient, startContainerIfClient),
		},
	}
}

// RemoveLinuxProducts is products remove flow for linux.
func RemoveLinuxProducts() flow.Flow {
	return flow.Flow{
		Name: "Products remove (linux)",
		Steps: []flow.Step{
			newStep("process flags", processedRemoveProductFlags, nil),
			newStep("start container if client", startContainerIfClient, stopContainerIfClient),
			newStep("validate", checkContainer, nil),
			newStep("remove products", removeProducts, nil),
			newStep("finalize", finalize, nil),
			newStep("stop if client", stopContainerIfClient, startContainerIfClient),
		},
	}
}
