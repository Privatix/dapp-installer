package command

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	dapputil "github.com/privatix/dappctrl/util"

	"github.com/privatix/dapp-installer/dapp"
	"github.com/privatix/dapp-installer/dbengine"
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

func processedRemoveFlags(d *dapp.Dapp) error {
	return processedWorkFlags(d, removeHelp)
}

func processedCommonFlags(d *dapp.Dapp, help string) error {
	h := flag.Bool("help", false, "Display dapp-installer help")
	config := flag.String("config", "", "Configuration file")
	role := flag.String("role", "", "Dapp user role")
	path := flag.String("workdir", "", "Dapp install directory")
	src := flag.String("source", "", "Dapp install source")

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

	return nil
}

func validatePath(d *dapp.Dapp) error {
	path, err := filepath.Abs(d.Path)
	if err != nil {
		return err
	}
	d.Path = filepath.ToSlash(strings.ToLower(path))
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
		ch := make(chan bool)
		defer close(ch)
		go util.InteractiveWorker("extracting dapp", ch)

		_, err := util.Unzip(path, d.Path)
		if err != nil {
			ch <- true
			return fmt.Errorf("failed to extract dapp: %v", err)
		}
		ch <- true
		fmt.Printf("\r%s\n", "dapp was successfully extracted")
	}
	fileName := filepath.Join(d.Path, d.Controller.EntryPoint)
	d.Version = util.DappCtrlVersion(fileName)

	return nil
}

func install(d *dapp.Dapp) error {
	if err := d.DBEngine.Install(d.Path); err != nil {
		return fmt.Errorf("failed to install dbengine: %v", err)
	}

	filePath := filepath.Join(d.Path, d.Controller.EntryPoint)
	if err := d.DBEngine.CreateDatabase(filePath); err != nil {
		return fmt.Errorf("failed to create db: %v", err)
	}

	if err := d.Configurate(); err != nil {
		return fmt.Errorf("failed to configurate dapp: %v", err)
	}

	return nil
}

func update(d *dapp.Dapp) error {
	oldDapp := *d

	b, dir := filepath.Split(oldDapp.Path)
	newPath := filepath.Join(b, dir+"_new")
	d.Path = newPath

	if err := extract(d); err != nil {
		return err
	}

	defer os.RemoveAll(newPath)

	version := util.ParseVersion(d.Version)
	if util.ParseVersion(oldDapp.Version) >= version {
		return fmt.Errorf(
			"dapp current version: %s, update is not required",
			oldDapp.Version)
	}

	// Update dapp core.
	if err := d.Update(&oldDapp); err != nil {
		oldDapp.DBEngine.Start(oldDapp.Path)
		oldDapp.Controller.Service.Start()
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

func stopServices(d *dapp.Dapp) error {
	done := make(chan bool)
	go d.Stop(done)

	select {
	case <-done:
		return nil
	case <-time.After(util.Timeout):
		return fmt.Errorf("failed to stop services: timeout expired")
	}
}

func removeServices(d *dapp.Dapp) error {
	if err := d.Controller.Service.Remove(); err != nil {
		return fmt.Errorf("failed to remove dappctrl service: %v", err)
	}

	if err := d.DBEngine.Remove(d.Path); err != nil {
		return fmt.Errorf("failed to remove db engine service: %v", err)
	}
	return nil
}

func removeDapp(d *dapp.Dapp) error {
	if err := d.Remove(); err != nil {
		return fmt.Errorf("failed to remove folder: %v", err)
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

	return nil
}

func processedWorkFlags(d *dapp.Dapp, help string) error {
	h := flag.Bool("help", false, "Display dapp-installer help")
	p := flag.String("workdir", ".", "Dapp install directory")

	flag.CommandLine.Parse(os.Args[2:])

	if *h {
		fmt.Println(help)
		os.Exit(0)
	}

	d.Path = *p
	return nil
}