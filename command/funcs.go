package command

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/privatix/dapp-installer/dapp"
	"github.com/privatix/dapp-installer/data"
	"github.com/privatix/dapp-installer/dbengine"
	"github.com/privatix/dapp-installer/env"
	"github.com/privatix/dapp-installer/util"
)

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

func prepare(d *dapp.Dapp) error {
	if err := util.ExecuteCommand("apt-get", "update"); err != nil {
		return fmt.Errorf("failed to prepare system: %v", err)
	}

	if err := util.ExecuteCommand("apt-get", "install",
		"systemd-container", "-y"); err != nil {
		return fmt.Errorf("failed to prepare system: %v", err)
	}

	return nil
}

func configureDapp(d *dapp.Dapp) error {
	engine := d.DBEngine
	engine.DB.Port, _ = util.FreePort(engine.DB.Host, engine.DB.Port)

	conf := filepath.Join(d.Path, "etc/postgresql/10/main/postgresql.conf")
	if err := dbengine.SetPort(conf, "5433", engine.DB.Port); err != nil {
		return fmt.Errorf("failed to configure db conf: %v", err)
	}

	if err := d.Configurate(); err != nil {
		return fmt.Errorf("failed to configure: %v", err)
	}
	return nil
}

func disableDaemons(d *dapp.Dapp) error {

	return nil
}

func createDatabase(d *dapp.Dapp) error {
	time.Sleep(10 * time.Second)

	file := filepath.Join(d.Path, d.Controller.EntryPoint)
	if err := d.DBEngine.CreateDatabase(file); err != nil {
		return fmt.Errorf("failed to create db: %v", err)
	}

	if err := d.DBEngine.Ping(); err != nil {
		return fmt.Errorf("failed to finalize: %v", err)
	}

	if err := writeVersion(d); err != nil {
		return fmt.Errorf("failed to write dapp version: %v", err)
	}

	return nil
}

func updateDatabase(d *dapp.Dapp) error {
	if err := d.DBEngine.Ping(); err != nil {
		return fmt.Errorf("failed to finalize: %v", err)
	}

	file := filepath.Join(d.Path, d.Controller.EntryPoint)
	if err := d.DBEngine.UpdateDatabase(file); err != nil {
		return fmt.Errorf("failed to update db: %v", err)
	}

	if err := writeVersion(d); err != nil {
		return fmt.Errorf("failed to write dapp version: %v", err)
	}

	return nil
}

func finalize(d *dapp.Dapp) error {
	if err := stopContainer(d); err != nil {
		return err
	}

	time.Sleep(10 * time.Second)

	return startContainer(d)
}

func removeBackup(d *dapp.Dapp) error {
	return os.RemoveAll(d.BackupPath)
}
