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
	"github.com/privatix/dappctrl/util/log"
)

// Dapp has a config for dapp core.
type Dapp struct {
	UserRole    string
	InstallPath string
	Source      string
	Controller  *InstallerEntity
	Gui         *InstallerEntity
	DBEngine    *dbengine.DBEngine
	TempPath    string
	BackupPath  string
	Version     string
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
		UserRole:    "agent",
		InstallPath: ".",
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

// Install installs a dapp core.
func (d *Dapp) Install(logger log.Logger) error {
	d.InstallPath, _ = filepath.Abs(d.InstallPath)
	if err := d.prepareToInstall(logger); err != nil {
		return err
	}

	// Install dbengine.
	if err := d.DBEngine.Install(d.InstallPath, logger); err != nil {
		return err
	}

	// Create DB.
	filePath := filepath.Join(d.InstallPath, d.Controller.EntryPoint)
	if err := d.DBEngine.CreateDatabase(filePath); err != nil {
		return err
	}

	// Configure dappctrl.
	return d.configurateController(logger)
}

// Update updates the dapp core.
func (d *Dapp) Update(oldDapp *Dapp, logger log.Logger) error {
	// Install db engine.
	ch := make(chan bool)
	defer close(ch)
	go util.InteractiveWorker("Upgrading Dapp", ch)

	// Stop services.
	done := make(chan bool)
	go oldDapp.stopServices(done)

	select {
	case <-done:
		logger.Info("services were successfully stopped")
	case <-time.After(util.Timeout):
		os.RemoveAll(d.InstallPath)
		ch <- true
		return errors.New("failed to stopped services. timeout expired")
	}

	// Merge with exist dapp.
	if err := d.merge(oldDapp); err != nil {
		os.RemoveAll(d.InstallPath)
		if len(oldDapp.BackupPath) > 0 {
			os.Rename(oldDapp.BackupPath, oldDapp.InstallPath)
		}
		ch <- true
		return err
	}
	logger.Info("dapp's were successfully merged")

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
	return nil
}

func (d *Dapp) modifyDappConfig(logger log.Logger) error {
	configFile := filepath.Join(d.InstallPath, d.Controller.Configuration)

	if err := setDynamicPorts(configFile, logger); err != nil {
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

	settings["Role"] = d.UserRole
	settings["FileLog.Filename"] = filepath.Join(d.InstallPath,
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
func (d *Dapp) Remove(logger log.Logger) error {
	ch := make(chan bool)
	defer close(ch)
	go util.InteractiveWorker("Removing Dapp", ch)

	if err := d.uninstallServices(logger); err != nil {
		ch <- true
		return err
	}

	d.removeFinalize(logger)

	if err := os.RemoveAll(d.InstallPath); err != nil {
		time.Sleep(10 * time.Second)
		ch <- true
		return os.RemoveAll(d.InstallPath)
	}
	ch <- true
	fmt.Printf("\r%s\n", "Dapp was successfully removed")

	return nil
}

func (d Dapp) controllerHash() string {
	hash := util.Hash(d.InstallPath)
	return fmt.Sprintf("dapp_ctrl_%s", hash)
}

// Exists returns existing dapp in the host.
func Exists(path string, logger log.Logger) (*Dapp, bool) {
	path, _ = filepath.Abs(path)

	dappCtrl := filepath.Join(path, "dappctrl")
	if _, err := os.Stat(dappCtrl); os.IsNotExist(err) {
		return nil, false
	}

	dappData := filepath.Join(path, "pgsql/data")
	if _, err := os.Stat(dappData); os.IsNotExist(err) {
		return nil, false
	}

	d := NewConfig()

	configFile := filepath.Join(path, d.Controller.Configuration)
	role, db, err := roleAndDBConnFromConfig(configFile)
	if err != nil {
		logger.Warn(fmt.Sprintf("failed to read config: %v", err))
		return nil, false
	}

	d.InstallPath = path
	d.Version = util.DappCtrlVersion(filepath.Join(path,
		d.Controller.EntryPoint))
	hash := d.controllerHash()
	d.Controller.Service.ID = hash
	d.Controller.Service.GUID = filepath.Join(dappCtrl, hash)

	d.UserRole, d.DBEngine.DB = role, db

	return d, true
}

// SetInstallPath sets value to dapp InstallPath params.
func (d *Dapp) SetInstallPath(path string) {
	d.InstallPath, _ = filepath.Abs(path)
}

func (d *Dapp) merge(s *Dapp) error {
	d.UserRole = s.UserRole
	d.DBEngine = s.DBEngine

	copyServiceWrapper(d, s)

	// Copy data.
	util.CopyDir(filepath.Join(s.InstallPath, "pgsql/data"),
		filepath.Join(d.InstallPath, "pgsql/data"))

	// Merge dappctrl config.
	dstConfig := filepath.Join(d.InstallPath, d.Controller.Configuration)
	srcConfig := filepath.Join(s.InstallPath, s.Controller.Configuration)
	if err := util.MergeJSONFile(dstConfig, srcConfig); err != nil {
		return err
	}

	s.BackupPath = util.RenamePath(s.InstallPath, "backup")
	if err := os.Rename(s.InstallPath, s.BackupPath); err != nil {
		return err
	}

	return os.Rename(d.InstallPath, s.InstallPath)
}

func (d *Dapp) stopServices(ch chan bool) {
	for {
		if !util.IsServiceStopped(d.Controller.Service.ID) {
			d.Controller.Service.Stop()
			time.Sleep(200 * time.Millisecond)
			continue
		}
		if !util.IsServiceStopped(dbengine.Hash(d.InstallPath)) {
			d.DBEngine.Stop(d.InstallPath)
			time.Sleep(200 * time.Millisecond)
			continue
		}
		break
	}
	ch <- true
}

func (d *Dapp) uninstallServices(logger log.Logger) error {
	// Stop services.
	done := make(chan bool)
	go d.stopServices(done)

	select {
	case <-done:
		logger.Info("services were successfully stopped")
	case <-time.After(util.Timeout):
		return errors.New("failed to stopped services.timeout expired")
	}

	// Remove services.
	if err := d.Controller.Service.Remove(); err != nil {
		return err
	}

	return d.DBEngine.Remove(d.InstallPath, logger)
}
