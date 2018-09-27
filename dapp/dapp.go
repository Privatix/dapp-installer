package dapp

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/privatix/dapp-installer/dbengine"
	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dapp-installer/windows"
	"github.com/privatix/dappctrl/util/log"
)

// Dapp has a config for dapp core.
type Dapp struct {
	UserRole    string
	InstallPath string
	Download    string
	Controller  *InstallerEntity
	Gui         *InstallerEntity
	DBEngine    *dbengine.DBEngine
	TempPath    string
	BackupPath  string
	Version     string
	Registry    *util.Registry
}

// InstallerEntity has a config for install entity.
type InstallerEntity struct {
	EntryPoint    string
	Configuration string
	Shortcuts     bool
	Service       *windows.Service
}

// NewConfig creates a default Dapp configuration.
func NewConfig() *Dapp {
	return &Dapp{
		DBEngine: dbengine.NewConfig(),
	}
}

// DownloadDapp downloads dapp and returns temporary download path.
func (d *Dapp) DownloadDapp() string {
	filePath := filepath.Join(d.TempPath, path.Base(d.Download))
	if err := util.DownloadFile(filePath, d.Download); err != nil {
		return ""
	}

	return filePath
}

// Install installs a dapp core.
func (d *Dapp) Install(logger log.Logger) error {
	// Install dbengine.
	err := d.DBEngine.Install(d.InstallPath, logger)
	if err != nil {
		return err
	}

	// Create DB.
	filePath := filepath.Join(d.InstallPath, d.Controller.EntryPoint)
	if err := d.DBEngine.CreateDatabase(filePath); err != nil {
		return err
	}

	// Configure dappctrl.
	if err := d.configurateController(logger); err != nil {
		return err
	}

	return d.installFinalize(logger)
}

// Update updates the dapp core.
func (d *Dapp) Update(oldDapp *Dapp, logger log.Logger) error {
	// Install db engine.
	ch := make(chan bool)
	defer close(ch)
	go util.InteractiveWorker("Upgrading Dapp", ch)

	// Stop services.
	oldDapp.Controller.Service.Stop()
	oldDapp.DBEngine.Stop(oldDapp.InstallPath)

	// Copy data.
	util.CopyDir(filepath.Join(oldDapp.InstallPath, "pgsql/data"),
		filepath.Join(d.InstallPath, "pgsql/data"))

	oldDapp.BackupPath = util.RenamePath(oldDapp.InstallPath, "backup")
	if err := os.Rename(oldDapp.InstallPath, oldDapp.BackupPath); err != nil {
		os.RemoveAll(d.InstallPath)
		ch <- true
		return err
	}
	logger.Info("existing dapp version was successfully backuped")

	if err := os.Rename(d.InstallPath, oldDapp.InstallPath); err != nil {
		os.RemoveAll(d.InstallPath)
		os.Rename(oldDapp.BackupPath, oldDapp.InstallPath)
		ch <- true
		return err
	}

	d.InstallPath = oldDapp.InstallPath

	// Start dbengine.
	d.DBEngine.Start(d.InstallPath)

	// Update DB schema.
	filePath := filepath.Join(d.InstallPath, d.Controller.EntryPoint)
	if err := d.DBEngine.UpdateDatabase(filePath); err != nil {
		d.DBEngine.Stop(d.InstallPath)
		os.RemoveAll(d.InstallPath)
		os.Rename(oldDapp.BackupPath, oldDapp.InstallPath)
		ch <- true
		return err
	}
	logger.Info("db migartions were successfully executed")

	// Configure dappctrl.
	if err := d.configurateController(logger); err != nil {
		d.DBEngine.Stop(d.InstallPath)
		os.RemoveAll(d.InstallPath)
		os.Rename(oldDapp.BackupPath, oldDapp.InstallPath)
		ch <- true
		return err
	}

	d.Controller.Service.Start()
	os.RemoveAll(oldDapp.BackupPath)
	ch <- true
	fmt.Printf("\r%s\n", "Dapp was successfully upgraded")
	return d.updateFinalize(logger)
}

func (d *Dapp) modifyDappConfig() error {
	configFile := filepath.Join(d.InstallPath, d.Controller.Configuration)
	read, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer read.Close()

	jsonMap := make(map[string]interface{})

	json.NewDecoder(read).Decode(&jsonMap)

	if db, ok := jsonMap["DB"].(map[string]interface{}); ok {
		if conn, ok := db["Conn"].(map[string]interface{}); ok {
			conn["user"] = d.DBEngine.DB.User
			if len(d.DBEngine.DB.Password) > 0 {
				conn["password"] = d.DBEngine.DB.Password
			}
			conn["port"] = d.DBEngine.DB.Port
			conn["dbname"] = d.DBEngine.DB.DBName
		}
	}

	if _, ok := jsonMap["Role"]; ok {
		jsonMap["Role"] = d.UserRole
	}

	write, err := os.Create(configFile)
	if err != nil {
		fmt.Println(err)
	}
	defer write.Close()

	return json.NewEncoder(write).Encode(jsonMap)
}

// Remove removes installed dapp core.
func (d *Dapp) Remove(logger log.Logger) error {
	d.Controller.Service.Uninstall()

	d.DBEngine.Remove(d.InstallPath, logger)
	d.removeFinalize(logger)

	if err := os.RemoveAll(d.InstallPath); err != nil {
		time.Sleep(10 * time.Second)
		return os.RemoveAll(d.InstallPath)
	}
	return nil
}
