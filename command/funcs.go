package command

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	dapputil "github.com/privatix/dappctrl/util"

	"github.com/privatix/dapp-installer/dapp"
	"github.com/privatix/dapp-installer/data"
	"github.com/privatix/dapp-installer/dbengine"
	"github.com/privatix/dapp-installer/env"
	"github.com/privatix/dapp-installer/product"
	"github.com/privatix/dapp-installer/util"
)

func processedRootFlags(printVersion func()) {
	v := flag.Bool("version", false, "Prints current dapp-installer version")

	flag.Parse()

	if *v {
		printVersion()
		os.Exit(0)
	}

	fmt.Printf(rootHelp)
}

func processedInstallFlags(d *dapp.Dapp) error {
	return processedCommonFlags(d, installHelp)
}

func processedUpdateFlags(d *dapp.Dapp) error {
	return processedCommonFlags(d, updateHelp)
}

func processedInstallProductFlags(d *dapp.Dapp) error {
	return processedCommonFlags(d, installProductHelp)
}

func processedUpdateProductFlags(d *dapp.Dapp) error {
	return processedCommonFlags(d, updateProductHelp)
}

func processedRemoveProductFlags(d *dapp.Dapp) error {
	return processedCommonFlags(d, removeProductHelp)
}

func processedRemoveFlags(d *dapp.Dapp) error {
	return processedWorkFlags(d, removeHelp)
}

func processedCommonFlags(d *dapp.Dapp, help string) error {
	h := flag.Bool("help", false, "Display dapp-installer help")
	config := flag.String("config", "", "Configuration file")
	role := flag.String("role", "", "Dapp user role")
	path := flag.String("workdir", "", "Dapp install directory")
	src := flag.String("source", "", "Dapp install source")
	core := flag.Bool("core", false, "Install only dapp core")
	product := flag.String("product", "", "Specific product")

	flag.Bool("verbose", false, "Display log to console output")

	flag.CommandLine.Parse(os.Args[2:])

	if *h {
		fmt.Println(help)
		os.Exit(0)
	}

	if len(*config) > 0 {
		if err := dapputil.ReadJSONFile(*config, &d); err != nil {
			return err
		}
	}

	if len(*role) > 0 {
		d.Role = *role
	}

	if len(*path) > 0 {
		d.Path = *path
	}

	if len(*src) > 0 {
		d.Source = *src
	}

	if len(*product) > 0 {
		d.Product = *product
	}

	d.OnlyCore = *core

	return nil
}

func validatePath(d *dapp.Dapp) error {
	path, err := filepath.Abs(d.Path)
	if err != nil {
		return err
	}
	d.Path = filepath.ToSlash(path)
	d.Tor.RootPath = d.Path
	return nil
}

func validateToInstall(d *dapp.Dapp) error {
	if err := validatePath(d); err != nil {
		return err
	}

	v := filepath.VolumeName(d.Path)

	if err := util.CheckSystemPrerequisites(v); err != nil {
		return fmt.Errorf("failed to validate: %v", err)
	}

	if err := d.Exists(); err == nil {
		return fmt.Errorf(
			"dapp is installed in '%s'. your dapp version: %v",
			d.Path, d.Version)
	}

	return nil
}

func initTemp(d *dapp.Dapp) error {
	v := filepath.VolumeName(d.Path)

	d.TempPath = util.TempPath(v)
	return nil
}

func removeTemp(d *dapp.Dapp) error {
	os.RemoveAll(d.TempPath)

	return nil
}

func extract(d *dapp.Dapp) error {
	path := d.Download()

	if _, err := os.Stat(d.Path); os.IsNotExist(err) {
		os.MkdirAll(d.Path, util.FullPermission)
	}

	if len(path) > 0 {
		if err := util.Unzip(path, d.Path); err != nil {
			return fmt.Errorf("failed to extract dapp: %v", err)
		}
	}
	fileName := filepath.Join(d.Path, d.Controller.EntryPoint)
	d.Version = util.DappCtrlVersion(fileName)

	return nil
}

func installDBEngine(d *dapp.Dapp) error {
	if err := d.DBEngine.Install(d.Path); err != nil {
		return fmt.Errorf("failed to install dbengine: %v", err)
	}

	filePath := filepath.Join(d.Path, d.Controller.EntryPoint)
	if err := d.DBEngine.CreateDatabase(filePath); err != nil {
		return fmt.Errorf("failed to create db: %v", err)
	}

	return nil
}

func removeDBEngine(d *dapp.Dapp) error {
	if err := d.DBEngine.Remove(d.Path); err != nil {
		return fmt.Errorf("failed to remove dbengine: %v", err)
	}

	return nil
}

func install(d *dapp.Dapp) error {
	if err := d.Configurate(); err != nil {
		return fmt.Errorf("failed to configure dapp: %v", err)
	}

	return nil
}

func remove(d *dapp.Dapp) error {
	if err := stopServices(d); err != nil {
		return err
	}

	return removeServices(d)
}

func update(d *dapp.Dapp) error {
	oldDapp := *d

	b, dir := filepath.Split(oldDapp.Path)
	newPath := filepath.Join(b, dir+"_new")
	d.Path = newPath

	if err := extract(d); err != nil {
		d.Path = oldDapp.Path
		return err
	}

	defer os.RemoveAll(newPath)

	version := util.ParseVersion(d.Version)
	if util.ParseVersion(oldDapp.Version) >= version {
		d.Path = oldDapp.Path
		return fmt.Errorf(
			"dapp current version: %s, update is not required",
			oldDapp.Version)
	}

	// Update dapp core.
	if err := d.Update(&oldDapp); err != nil {
		d.Path = oldDapp.Path
		return fmt.Errorf("failed to update dapp: %v", err)
	}

	return nil
}

func checkInstallation(d *dapp.Dapp) error {
	if err := validatePath(d); err != nil {
		return err
	}

	if err := d.Exists(); err != nil {
		return fmt.Errorf("dapp is not found at %s: %v", d.Path, err)
	}
	return nil
}

func startServices(d *dapp.Dapp) error {
	if err := d.DBEngine.Start(d.Path); err != nil {
		return fmt.Errorf("failed to start dbengine: %v", err)
	}
	if err := d.Controller.Service.Start(); err != nil {
		return fmt.Errorf("failed to start controller: %v", err)
	}
	return nil
}

func stopServices(d *dapp.Dapp) error {
	done := make(chan bool)
	go d.Stop(done)

	select {
	case <-done:
		return nil
	case <-time.After(util.TimeOutInSec(d.Timeout)):
		return fmt.Errorf("failed to stop services: timeout expired")
	}
}

func removeServices(d *dapp.Dapp) error {
	if d.Role == "agent" && runtime.GOOS == "windows" {
		// Removes firewall rule for payment reciever of dappctrl.
		args := []string{"-ExecutionPolicy", "Bypass", "-File",
			filepath.Join(d.Path, "dappctrl", "set-ctrlfirewall.ps1"),
			"-Remove", "-ServiceName", d.Controller.Service.ID}
		err := util.ExecuteCommand("powershell", args...)
		if err != nil {
			return fmt.Errorf("failed to remove firewall rule: %v", err)
		}
	}

	if err := d.Controller.Service.Remove(); err != nil {
		return fmt.Errorf("failed to remove dappctrl service: %v", err)
	}

	if err := d.DBEngine.Remove(d.Path); err != nil {
		return fmt.Errorf("failed to remove db engine service: %v", err)
	}
	return nil
}

func removeDapp(d *dapp.Dapp) error {
	if err := util.KillProcess(d.Path); err != nil {
		return fmt.Errorf("failed to kill process: %v", err)
	}

	// Removes unremoved services.
	d.Controller.Service.Remove()
	d.DBEngine.Remove(d.Path)
	d.Tor.Remove()

	if err := d.Remove(); err != nil {
		if err := util.SelfRemove(d.Path); err != nil {
			return fmt.Errorf("failed to remove folder: %v", err)
		}
	}
	return nil
}

func processedStatusFlags(d *dapp.Dapp) error {
	return processedWorkFlags(d, statusHelp)
}

func printStatus(d *dapp.Dapp) error {
	if err := validatePath(d); err != nil {
		return err
	}

	fmt.Printf("installation status checking at `%v`:\n", d.Path)

	v := filepath.VolumeName(d.Path)

	result := success

	if err := util.CheckSystemPrerequisites(v); err != nil {
		result = failed
	}

	fmt.Printf(status, "system prerequisites", result)

	if err := d.Exists(); err != nil {
		fmt.Printf(status, "dapp exists", failed)
		fmt.Printf(status, "error", err)
		return nil
	}

	fmt.Printf(status, "dapp exists", success)
	fmt.Printf(status, "dapp path", d.Path)
	fmt.Printf(status, "dapp version", d.Version)
	fmt.Printf(status, "dapp role", d.Role)
	fmt.Printf(status, "dapp ctrl", d.Controller.Service.ID)
	fmt.Printf(status, "dapp db", dbengine.Hash(d.Path))

	fmt.Printf(status, "db host", d.DBEngine.DB.Host)
	fmt.Printf(status, "db name", d.DBEngine.DB.DBName)
	fmt.Printf(status, "db port", d.DBEngine.DB.Port)
	fmt.Printf(status, "db user", d.DBEngine.DB.User)
	fmt.Printf(status, "db password", d.DBEngine.DB.Password)

	fmt.Printf(status, "dapp tor", d.Tor.ServiceName())
	fmt.Printf(status, "tor name", d.Tor.Hostname)

	return nil
}

func processedWorkFlags(d *dapp.Dapp, help string) error {
	h := flag.Bool("help", false, "Display dapp-installer help")
	p := flag.String("workdir", "", "Dapp install directory")

	flag.Bool("verbose", false, "Display log to console output")

	flag.CommandLine.Parse(os.Args[2:])

	if *h {
		fmt.Println(help)
		os.Exit(0)
	}

	if len(*p) == 0 {
		*p = filepath.Dir(os.Args[0])
	}
	d.Path = *p
	return nil
}

func installTor(d *dapp.Dapp) error {
	if err := d.Tor.Install(d.Role); err != nil {
		return fmt.Errorf("failed to install tor: %v", err)
	}
	return nil
}

func removeTor(d *dapp.Dapp) error {
	if err := d.Tor.Remove(); err != nil {
		return fmt.Errorf("failed to remove tor: %v", err)
	}

	return nil
}

func stopTor(d *dapp.Dapp) error {
	if err := d.Tor.Stop(); err != nil {
		return fmt.Errorf("failed to stop tor: %v", err)
	}

	return nil
}

func startTor(d *dapp.Dapp) error {
	if err := d.Tor.Start(); err != nil {
		return fmt.Errorf("failed to start tor: %v", err)
	}

	return nil
}

func writeVersion(d *dapp.Dapp) error {
	if err := data.WriteAppVersion(d.DBEngine.DB, d.Version); err != nil {
		return fmt.Errorf("failed to write app version: %v", err)
	}

	return nil
}

func writeEnvironmentVariable(d *dapp.Dapp) error {
	v := env.NewConfig()

	v.Role = d.Role
	v.Version = d.Version
	v.WorkDir = d.Path
	v.UserID = d.UserID
	v.Dapp.Controller = d.Controller.EntryPoint
	v.Dapp.Gui = d.Gui.EntryPoint
	v.Dapp.Service = d.Controller.Service.ID

	v.DB.Service = dbengine.Hash(d.Path)

	v.Tor.Service = d.Tor.ServiceName()

	path := filepath.Join(d.Path, envFile)
	return v.Write(path)
}

func installProducts(d *dapp.Dapp) error {
	if d.OnlyCore {
		return nil
	}

	conn := d.DBEngine.DB.ConnectionString()
	if err := product.Install(d.Role, d.Path, conn, d.Product); err != nil {
		return fmt.Errorf("failed to install products: %v", err)
	}

	return nil
}

func removeProducts(d *dapp.Dapp) error {
	if d.OnlyCore {
		return nil
	}

	if err := product.Remove(d.Role, d.Path, d.Product); err != nil {
		return fmt.Errorf("failed to remove products: %v", err)
	}

	return nil
}

func startProducts(d *dapp.Dapp) error {
	if err := product.Start(d.Role, d.Path, d.Product); err != nil {
		return fmt.Errorf("failed to start products: %v", err)
	}

	return nil
}

func updateProducts(d *dapp.Dapp) error {
	if len(d.Source) == 0 {
		return fmt.Errorf("product path not set")
	}

	err := os.Setenv("PRIVATIX_TEMP_PRODUCT", d.Source)
	if err != nil {
		return fmt.Errorf("failed to set env variables: %v", err)
	}

	defer os.Setenv("PRIVATIX_TEMP_PRODUCT", "")

	err = product.Update(d.Role, d.Path, d.Source, d.Product)
	if err != nil {
		return fmt.Errorf("failed to update products: %v", err)
	}

	util.CopyDir(filepath.Join(d.Source, "product"),
		filepath.Join(d.Path, "product"))

	return nil
}

func stopProducts(d *dapp.Dapp) error {
	if err := product.Stop(d.Role, d.Path, d.Product); err != nil {
		return fmt.Errorf("failed to stop products: %v", err)
	}

	return nil
}
