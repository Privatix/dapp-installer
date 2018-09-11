// +build windows

package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/privatix/dapp-installer/data"
	"github.com/privatix/dapp-installer/statik"

	"github.com/lxn/win"
	"github.com/privatix/dappctrl/util/log"
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

	// TODO(ubozov) check the drive specified for installation
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

// ExistingDappCtrlVersion returns existing dappctrl version.
func ExistingDappCtrlVersion(logger log.Logger) (string, bool) {
	v, err := getInstalledDappVersion()
	if err != nil {
		return "", false
	}
	return v, true
}

// InstallDappCtrl installs a dappctrl.
func InstallDappCtrl(installPath string, conf *DappCtrlConfig, dbConf *data.DB,
	logger log.Logger, ok bool) error {
	serviceName := installPath + "\\" + conf.Service.ID
	if ok {
		stopServiceWrapper(serviceName)
		removeServiceWrapper(serviceName)
	}
	logger.Info("createservicewrapper")
	if err := createServiceWrapper(installPath, conf.Service); err != nil {
		logger.Warn("ocurred error when create service wrapper:" +
			conf.Service.ID)
		return err
	}
	logger.Info("download dappctrl")
	if err := downloadController(installPath, conf, dbConf); err != nil {
		logger.Warn("ocurred error when downloaded files")
		return err
	}
	logger.Info("install service")
	if err := installServiceWrapper(serviceName); err != nil {
		logger.Warn("ocurred error when install service " + conf.Service.ID)
		return err
	}
	logger.Info("start service")
	if err := startServiceWrapper(serviceName); err != nil {
		logger.Warn("ocurred error when start service " + conf.Service.ID)
		return err
	}
	return nil
}

// RemoveDappCtrl removes installed dappctrl.
func RemoveDappCtrl(installPath string, conf *DappCtrlConfig) error {
	serviceName := installPath + "\\" + conf.Service.ID
	stopServiceWrapper(serviceName)
	removeServiceWrapper(serviceName)
	return os.RemoveAll(installPath)
}

func downloadController(installPath string, dappConf *DappCtrlConfig, dbConf *data.DB) error {
	fileName := path.Base(dappConf.DownloadDapp)
	err := downloadFile(installPath+"\\"+fileName, dappConf.DownloadDapp)
	if err != nil {
		return err
	}

	configFile := path.Base(dappConf.DownloadConfig)
	err = downloadFile(installPath+"\\"+configFile, dappConf.DownloadConfig)
	if err != nil {
		return err
	}
	return modifyDappConfig(installPath+"\\"+configFile, dbConf)
}

func modifyDappConfig(configFile string, dbConf *data.DB) error {
	read, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer read.Close()

	jsonMap := make(map[string]interface{})

	json.NewDecoder(read).Decode(&jsonMap)

	if db, ok := jsonMap["DB"].(map[string]interface{}); ok {
		if conn, ok := db["Conn"].(map[string]interface{}); ok {
			conn["user"] = dbConf.User
			conn["password"] = dbConf.Password
			conn["port"] = dbConf.Port
			conn["dbname"] = dbConf.DBName
		}
	}

	write, err := os.Create(configFile)
	if err != nil {
		fmt.Println(err)
	}
	defer write.Close()

	return json.NewEncoder(write).Encode(jsonMap)
}

func createServiceWrapper(path string, service *serviceConfig) error {
	// create winservice wrapper file
	bytes, err := statik.ReadFile("/wrapper/winsvc.exe")
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0644)
	}

	err = ioutil.WriteFile(path+"\\"+service.ID+".exe", bytes, 0644)
	if err != nil {
		return err
	}
	// create winservice wrapper config file
	service.Command = path + "\\" + service.Command
	for i, arg := range service.Args {
		if strings.HasSuffix(arg, ".json") {
			service.Args[i] = path + "\\" + arg
		}
	}
	bytes, err = json.Marshal(service)
	if err != nil {
		fmt.Println("error:", err)
	}
	return ioutil.WriteFile(path+"\\"+service.ID+".config.json", bytes, 0644)
}

func installServiceWrapper(wrapperName string) error {
	return ExecuteCommand(wrapperName, []string{"install"})
}

func startServiceWrapper(wrapperName string) error {
	return ExecuteCommand(wrapperName, []string{"start"})
}

func stopServiceWrapper(wrapperName string) error {
	return ExecuteCommand(wrapperName, []string{"stop"})
}

func removeServiceWrapper(wrapperName string) error {
	return ExecuteCommand(wrapperName, []string{"remove"})
}
