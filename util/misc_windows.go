// +build windows

package util

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/Privatix/dappctrl/util/log"
	"github.com/lxn/win"
	"golang.org/x/sys/windows/registry"
)

// MinWindowsVersion is supported min windows version (Windows7 and newer)
const MinWindowsVersion byte = 6

// MinDBEngineVersion is supported min DB Engine version
const MinDBEngineVersion int = 9

// WinRegDBEngine64 is a registry path to 64 bit DBEngine installations
const WinRegDBEngine64 string = `SOFTWARE\PostgreSQL\Installations\`

// WinRegDBEngine32 is a registry path to 32 bit DBEngine installations
const WinRegDBEngine32 string = `SOFTWARE\WOW6432Node\PostgreSQL\Installations\`

// WinRegInstalledDapp is a registry path to dapp installations
const WinRegInstalledDapp string = `SOFTWARE\Privatix\Dapp\`

// WinRegUninstallDapp is a registry path to dapp uninstallations
const WinRegUninstallDapp string = `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\Privatix Dapp\`

// CheckSystemPrerequisites does checked system to prerequisites.
func CheckSystemPrerequisites(logger log.Logger) bool {
	if runtime.GOOS != "windows" {
		logger.Warn("software install only to Windows platform")
		return false
	}

	if !checkWindowsVersion() {
		logger.Warn("Windows version does not meet the requirements")
		return false
	}

	if !checkMemory() {
		logger.Warn("RAM does not meet the requirements")
		return false
	}

	if !checkStorage() {
		logger.Warn("available size of disk does not meet the requirements")
		return false
	}

	return true
}

func checkWindowsVersion() bool {
	v := win.GetVersion()
	major := byte(v)
	return major >= MinWindowsVersion
}

func checkStorage() bool {
	h := syscall.MustLoadDLL("kernel32.dll")
	c := h.MustFindProc("GetDiskFreeSpaceExW")
	lpFreeBytesAvailable := uint64(0)
	u := utf16.Encode([]rune("\\"))
	c.Call(uintptr(unsafe.Pointer(&u[0])),
		uintptr(unsafe.Pointer(&lpFreeBytesAvailable)))

	return lpFreeBytesAvailable >= MinAvailableDiskSize
}

func checkMemory() bool {
	var totalMemoryInKilobytes uint64
	win.GetPhysicallyInstalledSystemMemory(&totalMemoryInKilobytes)
	return totalMemoryInKilobytes*1024 > MinMemorySize
}

// ExistingDBEnginePort returns existing db engine port number.
func ExistingDBEnginePort(logger log.Logger) (int, bool) {
	if p, ok := getDBEngineServicePort(WinRegDBEngine64); ok {
		return p, ok
	}
	return getDBEngineServicePort(WinRegDBEngine32)
}

func getDBEngineServicePort(keyPath string) (int, bool) {
	path, ok := checkDBEngineVersion(registry.LOCAL_MACHINE, keyPath)
	if ok {
		s := strings.Replace(keyPath, "Installations", "Services", 1)
		path = s + path
		return getServicePort(registry.LOCAL_MACHINE, path), true
	}
	return 0, false
}

// InstallDBEngine is installing DB engine.
func InstallDBEngine(conf *DBEngine, logger log.Logger) error {
	fileName := path.Base(conf.Download)
	if err := downloadFile(fileName, conf.Download); err != nil {
		logger.Warn("ocurred error when downloded file from " + conf.Download)
		return err
	}
	logger.Info("file successfully downloaded")

	// install db engine
	ch := make(chan bool)
	defer close(ch)
	go interactiveWorker("Installation DB Engine", ch)

	if err := exec.Command(fileName,
		generateDBEngineInstallParams(conf)...).Run(); err != nil {
		ch <- true
		fmt.Printf("\r%s\n", "Ocurred error when install DB Engine")
		logger.Warn("ocurred error when install dbengine")
		return err
	}
	logger.Info("dbengine successfully installed")

	ch <- true
	fmt.Printf("\r%s\n", "DB Engine successfully installed")

	for _, c := range conf.Copy {
		fmt.Println(c.From, c.To)
		fileName := path.Base(c.From)
		if err := downloadFile(c.To+"\\"+fileName, c.From); err != nil {
			logger.Warn("ocurred error when downloded file from " + c.From)
			return err
		}
	}

	// start db engine service
	if err := startService(conf.ServiceName); err != nil {
		logger.Warn("ocurred error when start dbengine service")
		return err
	}

	logger.Info("dbengine service successfully started")
	return nil
}

func startService(service string) error {
	checkServiceCmd := exec.Command("sc", "queryex", service)

	var checkServiceStdOut bytes.Buffer
	checkServiceCmd.Stdout = &checkServiceStdOut

	if err := checkServiceCmd.Run(); err != nil {
		return err
	}

	// service is running
	if strings.Contains(checkServiceStdOut.String(), "RUNNING") {
		return nil
	}

	// trying start service
	return exec.Command("net", "start", service).Run()
}

// ExecuteCommand does executing file.
func ExecuteCommand(filename string, args []string) error {
	cmd := exec.Command(filename, args...)

	return cmd.Run()
}
