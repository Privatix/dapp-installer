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

// CheckSystemPrerequisites does checked system to prerequisites.
func CheckSystemPrerequisites(log log.Logger) bool {
	if runtime.GOOS != "windows" {
		log.Warn("software install only to Windows platform")
		return false
	}

	if !checkWindowsVersion() {
		log.Warn("Windows version does not meet the requirements")
		return false
	}

	if !checkMemory() {
		log.Warn("RAM does not meet the requirements")
		return false
	}

	if !checkStorage() {
		log.Warn("available size of disk does not meet the requirements")
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

// DBEngineExists is checking to install DB engine.
func DBEngineExists(log log.Logger) bool {
	return checkDBEngineVersion(registry.LOCAL_MACHINE, WinRegDBEngine64) ||
		checkDBEngineVersion(registry.LOCAL_MACHINE, WinRegDBEngine32)
}

// InstallDBEngine is installing DB engine.
func InstallDBEngine(conf *DBEngine, log log.Logger) error {
	fileName := path.Base(conf.Download)
	if err := downloadFile(fileName, conf.Download); err != nil {
		log.Warn("ocurred error when downloded file from " + conf.Download)
		return err
	}
	log.Info("file successfully downloaded")

	// install db engine
	ch := make(chan bool)
	defer close(ch)
	go interactiveWorker("Installation DB Engine", ch)

	if err := exec.Command(fileName,
		generateDBEngineInstallParams(conf)...).Run(); err != nil {
		ch <- true
		fmt.Printf("\r%s\n", "Ocurred error when install DB Engine")
		log.Warn("ocurred error when install dbengine")
		return err
	}
	log.Info("dbengine successfully installed")

	ch <- true
	fmt.Printf("\r%s\n", "DB Engine successfully installed")

	for _, c := range conf.Copy {
		fmt.Println(c.From, c.To)
		fileName := path.Base(c.From)
		if err := downloadFile(c.To+"\\"+fileName, c.From); err != nil {
			log.Warn("ocurred error when downloded file from " + c.From)
			return err
		}
	}

	// start db engine service
	if err := startService(conf.ServiceName); err != nil {
		log.Warn("ocurred error when start dbengine service")
		return err
	}

	log.Info("dbengine service successfully started")
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
