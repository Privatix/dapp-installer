package dapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/privatix/dapp-installer/dbengine"
	"github.com/privatix/dapp-installer/util"
)

// Dapp has a config for dapp core.
type Dapp struct {
	Role       string
	Path       string
	Source     string
	Controller *InstallerEntity
	Gui        *InstallerEntity
	DBEngine   *dbengine.DBEngine
	TempPath   string
	BackupPath string
	Version    string
}

// InstallerEntity has a config for install entity.
type InstallerEntity struct {
	EntryPoint    string
	Configuration string
	Shortcuts     bool
	Service       *service
	Settings      map[string]interface{}
}

// NewConfig creates a default Dapp configuration.
func NewConfig() *Dapp {
	return &Dapp{
		Role: "agent",
		Path: ".",
		Controller: &InstallerEntity{
			EntryPoint:    "dappctrl/dappctrl",
			Configuration: "dappctrl/dappctrl.config.json",
			Service:       newService(),
			Settings:      make(map[string]interface{}),
		},
		Gui: &InstallerEntity{
			EntryPoint: "dappgui/dapp-gui",
			Shortcuts:  true,
		},
		DBEngine: dbengine.NewConfig(),
	}
}

// Download downloads dapp and returns temporary download path.
func (d *Dapp) Download() string {
	if !util.IsURL(d.Source) {
		return d.Source
	}

	filePath := filepath.Join(d.TempPath, path.Base(d.Source))
	if err := util.DownloadFile(filePath, d.Source); err != nil {
		return ""
	}

	return filePath
}

// Update updates the dapp core.
func (d *Dapp) Update(oldDapp *Dapp) error {
	// Install db engine.
	ch := make(chan bool)
	defer close(ch)
	go util.InteractiveWorker("upgrading dapp", ch)

	// Stop services.
	done := make(chan bool)
	go oldDapp.Stop(done)

	select {
	case <-done:

	case <-time.After(util.Timeout):
		os.RemoveAll(d.Path)
		ch <- true
		return errors.New("failed to stopped services. timeout expired")
	}

	// Merge with exist dapp.
	if err := d.merge(oldDapp); err != nil {
		os.RemoveAll(d.Path)
		if len(oldDapp.BackupPath) > 0 {
			os.Rename(oldDapp.BackupPath, oldDapp.Path)
		}
		ch <- true
		return err
	}

	d.Path = oldDapp.Path

	// Start dbengine.
	d.DBEngine.Start(d.Path)

	// Update DB schema.
	filePath := filepath.Join(d.Path, d.Controller.EntryPoint)
	if err := d.DBEngine.UpdateDatabase(filePath); err != nil {
		d.DBEngine.Stop(d.Path)
		os.RemoveAll(d.Path)
		os.Rename(oldDapp.BackupPath, oldDapp.Path)
		ch <- true
		return err
	}

	// Configure dappctrl.
	if err := d.Configurate(); err != nil {
		d.DBEngine.Stop(d.Path)
		os.RemoveAll(d.Path)
		os.Rename(oldDapp.BackupPath, oldDapp.Path)
		ch <- true
		return err
	}

	d.Controller.Service.Start()
	os.RemoveAll(oldDapp.BackupPath)
	ch <- true
	fmt.Printf("\r%s\n", "dapp was successfully upgraded")
	return nil
}

func (d *Dapp) modifyDappConfig() error {
	configFile := filepath.Join(d.Path, d.Controller.Configuration)

	if err := setDynamicPorts(configFile); err != nil {
		return err
	}

	read, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer read.Close()

	jsonMap := make(map[string]interface{})

	json.NewDecoder(read).Decode(&jsonMap)

	settings := d.Controller.Settings

	settings["Role"] = d.Role
	settings["FileLog.Filename"] = filepath.Join(d.Path,
		"log/dappctrl-%Y-%m-%d.log")
	settings["DB.Conn.user"] = d.DBEngine.DB.User
	settings["DB.Conn.port"] = d.DBEngine.DB.Port
	settings["DB.Conn.dbname"] = d.DBEngine.DB.DBName
	if len(d.DBEngine.DB.Password) > 0 {
		settings["DB.Conn.password"] = d.DBEngine.DB.Password
	}

	if err := setConfigurationValues(jsonMap, settings); err != nil {
		return err
	}

	write, err := os.Create(configFile)
	if err != nil {
		return err
	}
	defer write.Close()

	return json.NewEncoder(write).Encode(jsonMap)
}

// Remove removes installed dapp core.
func (d *Dapp) Remove() error {
	d.removeFinalize()

	if err := os.RemoveAll(d.Path); err != nil {
		time.Sleep(10 * time.Second)
		return os.RemoveAll(d.Path)
	}
	return nil
}

func (d Dapp) controllerHash() string {
	hash := util.Hash(d.Path)
	return fmt.Sprintf("dapp_ctrl_%s", hash)
}

// Exists returns existing dapp in the host.
func (d Dapp) Exists() error {
	dappCtrl := filepath.Join(d.Path, "dappctrl")
	if _, err := os.Stat(dappCtrl); os.IsNotExist(err) {
		return err
	}

	dappData := filepath.Join(d.Path, "pgsql/data")
	if _, err := os.Stat(dappData); os.IsNotExist(err) {
		return err
	}

	configFile := filepath.Join(d.Path, d.Controller.Configuration)
	role, db, err := roleAndDBConnFromConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config: %v", err)
	}

	d.Version = util.DappCtrlVersion(filepath.Join(d.Path,
		d.Controller.EntryPoint))
	hash := d.controllerHash()
	d.Controller.Service.ID = hash
	d.Controller.Service.GUID = filepath.Join(dappCtrl, hash)

	d.Role = role
	d.DBEngine.DB = db

	return nil
}

func (d *Dapp) merge(s *Dapp) error {
	d.Role = s.Role
	d.DBEngine = s.DBEngine

	copyServiceWrapper(d, s)

	// Copy data.
	util.CopyDir(filepath.Join(s.Path, "pgsql/data"),
		filepath.Join(d.Path, "pgsql/data"))

	// Merge dappctrl config.
	dstConfig := filepath.Join(d.Path, d.Controller.Configuration)
	srcConfig := filepath.Join(s.Path, s.Controller.Configuration)
	if err := util.MergeJSONFile(dstConfig, srcConfig); err != nil {
		return err
	}

	s.BackupPath = util.RenamePath(s.Path, "backup")
	if err := os.Rename(s.Path, s.BackupPath); err != nil {
		return err
	}

	return os.Rename(d.Path, s.Path)
}

// Stop stops dappctrl and debengine service.
func (d *Dapp) Stop(ch chan bool) {
	for {
		if !util.IsServiceStopped(d.Controller.Service.ID) {
			d.Controller.Service.Stop()
			time.Sleep(200 * time.Millisecond)
			continue
		}
		if !util.IsServiceStopped(dbengine.Hash(d.Path)) {
			d.DBEngine.Stop(d.Path)
			time.Sleep(200 * time.Millisecond)
			continue
		}
		break
	}
	ch <- true
}
