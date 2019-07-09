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
			newStep("extract", extractAndUpdateVersion, removeDapp),
			newStep("install tor", installTor, removeTor),
			newStep("start tor", startTor, stopTor),
			newStep("install dbengine", installDBEngine, removeDBEngine),
			newStep("install", install, remove),
			newStep("install products", installProducts, removeProducts),
			newStep("write version", writeVersion, nil),
			newStep("update sendremote setting", updateSendRemote, nil),
			newStep("write env", writeEnvironmentVariable, nil),
			newStep("remove temp", removeTemp, nil),
		},
	}
}

// Update is core update flow.
func Update() flow.Flow {
	if runtime.GOOS == "linux" {
		return UpdateLinux()
	}
	return flow.Flow{
		Name: "Update",
		Steps: []flow.Step{
			newStep("process flags", processedUpdateFlags, nil),
			newStep("validate", checkInstallation, nil),
			newStep("init temp", initTemp, removeTemp),
			newStep("stop tor", stopTor, startTor),
			newStep("stop services", stopServices, startServices),
			newStep("stop product services", stopProducts, nil),
			newStep("update", update, startProducts),
			newStep("write version", writeVersion, nil),
			newStep("write env", writeEnvironmentVariable, nil),
			newStep("start tor", startTor, nil),
			newStep("remove temp", removeTemp, nil),
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
			newStep("process flags", processedRemoveFlags, nil),
			newStep("validate", checkInstallation, nil),
			newStep("stop services", stopServices, nil),
			newStep("stop tor", stopTor, nil),
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
		},
	}
}

// UpdateProducts is products update flow.
func UpdateProducts() flow.Flow {
	if runtime.GOOS == "linux" {
		return UpdateLinuxProducts()
	}
	return flow.Flow{
		Name: "Products update",
		Steps: []flow.Step{
			newStep("process flags", processedUpdateProductFlags, nil),
			newStep("validate", checkInstallation, nil),
			newStep("update products", updateProducts, nil),
			newStep("start products", startProducts, nil),
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
			newStep("finalize", finalize, nil),
			newStep("remove temp", removeTemp, nil),
		},
	}
}

// UpdateLinux is core update flow for linux.
func UpdateLinux() flow.Flow {
	return flow.Flow{
		Name: "Update linux",
		Steps: []flow.Step{
			newStep("process flags", processedUpdateFlags, nil),
			newStep("validate", checkContainer, nil),
			newStep("init temp", initTemp, removeTemp),
			newStep("stop", disableAndStopContainer, enableAndStartContainer),
			newStep("update", updateContainer, restoreContainer),
			newStep("disable daemons", disableDaemons, nil),
			newStep("start", enableAndStartContainer, disableAndStopContainer),
			newStep("update database", updateDatabase, nil),
			newStep("enable daemons", enableDaemons, nil),
			newStep("finalize", finalize, nil),
			newStep("remove temp", removeTemp, nil),
			newStep("remove backup", removeBackup, nil),
		},
	}
}

// RemoveLinux is remove flow for linux.
func RemoveLinux() flow.Flow {
	return flow.Flow{
		Name: "removeLinux",
		Steps: []flow.Step{
			newStep("process flags", processedRemoveFlags, nil),
			newStep("validate", checkContainer, nil),
			newStep("stop", stopContainer, nil),
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
			newStep("validate", checkContainer, nil),
			newStep("install products", installProducts, removeProducts),
			newStep("finalize", finalize, nil),
		},
	}
}

// UpdateLinuxProducts is products update flow for linux.
func UpdateLinuxProducts() flow.Flow {
	return flow.Flow{
		Name: "Products update (linux)",
		Steps: []flow.Step{
			newStep("process flags", processedUpdateProductFlags, nil),
			newStep("validate", checkContainer, nil),
			newStep("update products", updateProducts, nil),
			newStep("start products", startProducts, nil),
			newStep("finalize", finalize, nil),
		},
	}
}

// RemoveLinuxProducts is products remove flow for linux.
func RemoveLinuxProducts() flow.Flow {
	return flow.Flow{
		Name: "Products remove (linux)",
		Steps: []flow.Step{
			newStep("process flags", processedRemoveProductFlags, nil),
			newStep("validate", checkContainer, nil),
			newStep("remove products", removeProducts, nil),
			newStep("finalize", finalize, nil),
		},
	}
}
