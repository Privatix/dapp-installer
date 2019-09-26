package flows

import (
	"context"
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

func validatePathAndSetAsRootPathForTorIfValid(d *dapp.Dapp) error {
	path, err := filepath.Abs(d.Path)
	if err != nil {
		return err
	}
	d.Path = filepath.ToSlash(path)
	d.Tor.RootPath = d.Path
	return nil
}

func validateToInstall(d *dapp.Dapp) error {
	if err := validatePathAndSetAsRootPathForTorIfValid(d); err != nil {
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

func extractAndUpdateVersion(d *dapp.Dapp) error {
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
	if err := d.Configure(); err != nil {
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

func checkInstallation(d *dapp.Dapp) (err error) {
	err = validatePathAndSetAsRootPathForTorIfValid(d)
	if err != nil {
		return err
	}
	d.Controller.Service.ID = d.ControllerHash()
	d.Controller.Service.GUID = filepath.Join(d.Path, "dappctrl", d.Controller.Service.ID)

	err = startServicesIfClient(d)
	defer func() {
		if err != nil {
			stopServicesIfClient(d)
		}
	}()
	if err != nil {
		return err
	}

	err = d.Exists()
	if err != nil {
		return fmt.Errorf("dapp is not found at %s: %v", d.Path, err)
	}
	return nil
}

func startServices(d *dapp.Dapp) error {
	d.Controller.Service.ID = d.ControllerHash()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return d.Start(ctx, "")
}

func stopServices(d *dapp.Dapp) error {
	ctx, cancel := context.WithTimeout(context.Background(), util.TimeOutInSec(d.Timeout))
	defer cancel()
	return d.Stop(ctx, "")
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
	if err := validatePathAndSetAsRootPathForTorIfValid(d); err != nil {
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

func updateSendRemote(d *dapp.Dapp) error {
	if err := data.UpdateSetting(d.DBEngine.DB, "error.sendremote", fmt.Sprint(d.SendRemote)); err != nil {
		return fmt.Errorf("failed to update error.sendremote setting: %v", err)
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

	if err := d.Configure(); err != nil {
		return fmt.Errorf("failed to configure: %v", err)
	}
	return nil
}

func disableDaemons(d *dapp.Dapp) error {
	src := filepath.Join(d.Path, dappCtrlDaemonOnLinux)
	dst := filepath.Join(d.Path, dappCtrlDaemonOnLinux+"temp")

	return util.ExecuteCommand("mv", src, dst)
}

func enableDaemons(d *dapp.Dapp) error {
	src := filepath.Join(d.Path, dappCtrlDaemonOnLinux+"temp")
	dst := filepath.Join(d.Path, dappCtrlDaemonOnLinux)

	return util.ExecuteCommand("mv", src, dst)
}

func createDatabase(d *dapp.Dapp) error {
	file := filepath.Join(d.Path, d.Controller.EntryPoint)
	if err := d.DBEngine.CreateDatabase(file); err != nil {
		return fmt.Errorf("failed to create db: %v", err)
	}

	if err := d.DBEngine.Ping(); err != nil {
		return fmt.Errorf("failed to finalize: %v", err)
	}

	if err := updateSendRemote(d); err != nil {
		return err
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

	return nil
}

func finalize(d *dapp.Dapp) error {
	ctx, cancel := context.WithTimeout(context.Background(),
		util.TimeOutInSec(d.Timeout))
	defer cancel()
	return util.RetryTillSucceed(ctx,
		func() error {
			err := restartContainer(d)
			if err != nil {
				stopContainer(d)
				time.Sleep(time.Second)
			}
			return err
		})
}

func removeBackup(d *dapp.Dapp) error {
	return os.RemoveAll(d.BackupPath)
}

func stopTorIfClient(d *dapp.Dapp) error {
	return runIfClient(d, stopTor)
}

func startTorIfClient(d *dapp.Dapp) error {
	return runIfClient(d, startTor)
}

func stopServicesIfClient(d *dapp.Dapp) error {
	return runIfClient(d, stopServices)
}

func startServicesIfClient(d *dapp.Dapp) error {
	err := runIfClient(d, startServices)
	return err
}

func stopProductsIfClient(d *dapp.Dapp) error {
	return runIfClient(d, stopProducts)
}

func startProductsIfClient(d *dapp.Dapp) error {
	return runIfClient(d, startProducts)
}

func stopContainerIfClient(d *dapp.Dapp) error {
	return runIfClient(d, stopContainer)
}

func startContainerIfClient(d *dapp.Dapp) error {
	return runIfClient(d, startContainer)
}

func runIfClient(d *dapp.Dapp, f func(*dapp.Dapp) error) error {
	if d.Role == "agent" {
		return nil
	}
	return f(d)
}
