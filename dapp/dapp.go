package dapp

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"

	"github.com/privatix/dapp-installer/data"
	"github.com/privatix/dapp-installer/dbengine"
	"github.com/privatix/dapp-installer/product"
	"github.com/privatix/dapp-installer/tor"
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
	Tor        *tor.Tor
	UserID     string
	Timeout    uint64 // in seconds
	OnlyCore   bool
	Product    string
	Verbose    bool
}

// InstallerEntity has a config for install entity.
type InstallerEntity struct {
	DisplayName   string
	EntryPoint    string
	Configuration string
	Service       *service
	Settings      map[string]interface{}
}

// NewDapp creates a default Dapp configuration.
func NewDapp() *Dapp {
	gc := "dappgui/resources/app/settings.json"

	if runtime.GOOS == "darwin" {
		gc = "dappgui/dapp-gui.app/Contents/Resources/app/settings.json"
	}

	return &Dapp{
		Role: "agent",
		Path: "./agent",
		Controller: &InstallerEntity{
			EntryPoint:    "dappctrl/dappctrl",
			Configuration: "dappctrl/dappctrl.config.json",
			Service:       newService(),
			Settings:      make(map[string]interface{}),
		},
		Gui: &InstallerEntity{
			DisplayName:   "Privatix",
			EntryPoint:    "dappgui/dapp-gui",
			Configuration: gc,
			Settings:      make(map[string]interface{}),
		},
		DBEngine: dbengine.NewConfig(),
		Tor:      tor.NewTor(),
		Timeout:  300,
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
	// Update products.
	if err := product.Update(d.Role, oldDapp.Path, d.Path,
		d.Product); err != nil {
		return fmt.Errorf("failed to update products: %v", err)
	}

	// Merge with exist dapp.
	if err := d.merge(oldDapp); err != nil {
		os.RemoveAll(d.Path)
		if len(oldDapp.BackupPath) > 0 {
			os.Rename(oldDapp.BackupPath, oldDapp.Path)
		}
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
		return err
	}

	// Configure dappctrl.
	if err := d.Configurate(); err != nil {
		d.DBEngine.Stop(d.Path)
		os.RemoveAll(d.Path)
		os.Rename(oldDapp.BackupPath, oldDapp.Path)
		return err
	}

	d.Controller.Service.Start()
	os.RemoveAll(oldDapp.BackupPath)
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

	d.UserID, err = machineid.ProtectedID("privatix")
	if err != nil {
		return err
	}
	settings := d.Controller.Settings

	settings[role] = d.Role
	settings[torHostname] = d.Tor.Hostname
	settings[torSocksListener] = d.Tor.SocksPort
	settings["SOMCServer.Addr"] = fmt.Sprintf("localhost:%v",
		d.Tor.TargetPort)
	settings["Report.userid"] = d.UserID

	if runtime.GOOS != "linux" {
		settings["FileLog.Filename"] = filepath.Join(d.Path,
			"log/dappctrl-%Y-%m-%d.log")
	}
	settings["DB.Conn.host"] = d.DBEngine.DB.Host
	settings["DB.Conn.user"] = d.DBEngine.DB.User
	settings["DB.Conn.port"] = d.DBEngine.DB.Port
	settings["DB.Conn.dbname"] = d.DBEngine.DB.DBName
	if len(d.DBEngine.DB.Password) > 0 {
		settings["DB.Conn.password"] = d.DBEngine.DB.Password
	}

	if strings.EqualFold(d.Role, "agent") {
		ip, err := externalIP()
		if err != nil {
			return err
		}
		addr := jsonMap["PayServer"].(map[string]interface{})["Addr"].(string)
		payServerAddr := strings.Replace(addr, "0.0.0.0", ip, 1)
		settings["PayAddress"] = strings.Replace(jsonMap["PayAddress"].(string),
			addr, payServerAddr, 1)

		if err := createFirewallRule(d, addr); err != nil {
			return err
		}
	}

	if err := setConfigurationValues(jsonMap, settings); err != nil {
		return err
	}

	addr := jsonMap["UI"].(map[string]interface{})["Addr"].(string)
	d.Gui.Settings["wsEndpoint"] = fmt.Sprintf("ws://%s/ws", addr)
	d.Gui.Settings["bugsnag.userid"] = d.UserID
	if err := d.setUIConfig(); err != nil {
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
	if err := clearGuiStorage(); err != nil {
		return err
	}

	if err := os.RemoveAll(d.Path); err != nil {
		time.Sleep(10 * time.Second)
		return os.RemoveAll(d.Path)
	}
	return nil
}

// Exists returns existing dapp in the host.
func (d *Dapp) Exists() error {
	dappCtrl := filepath.Join(d.Path, "dappctrl")
	if _, err := os.Stat(dappCtrl); os.IsNotExist(err) {
		return err
	}

	dappData := filepath.Join(d.Path, "pgsql", "data")
	if _, err := os.Stat(dappData); os.IsNotExist(err) {
		return err
	}

	if err := d.fromConfig(); err != nil {
		return fmt.Errorf("failed to read config: %v", err)
	}

	version, ok := data.ReadAppVersion(d.DBEngine.DB)
	if !ok {
		return fmt.Errorf("failed to read app version")
	}

	d.Version = version
	hash := d.controllerHash()
	d.Controller.Service.ID = hash
	d.Controller.Service.GUID = filepath.Join(dappCtrl, hash)

	return nil
}

func (d *Dapp) merge(s *Dapp) error {
	d.Role = s.Role
	d.DBEngine = s.DBEngine

	copyServiceWrapper(d, s)

	// Copy data.
	util.CopyDir(filepath.Join(s.Path, "pgsql", "data"),
		filepath.Join(d.Path, "pgsql", "data"))
	// Copy tor settings.
	util.CopyDir(filepath.Join(s.Path, "tor", "settings"),
		filepath.Join(d.Path, "tor", "settings"))
	util.CopyDir(filepath.Join(s.Path, s.Tor.HiddenServiceDir),
		filepath.Join(d.Path, d.Tor.HiddenServiceDir))

	// Merge dappctrl config.
	dstConfig := filepath.Join(d.Path, d.Controller.Configuration)
	srcConfig := filepath.Join(s.Path, s.Controller.Configuration)
	if err := util.MergeJSONFile(dstConfig, srcConfig, "PSCAddrHex"); err != nil {
		return err
	}

	// Merge dappgui config.
	dstConfig = filepath.Join(d.Path, d.Gui.Configuration)
	srcConfig = filepath.Join(s.Path, s.Gui.Configuration)
	if err := util.MergeJSONFile(dstConfig, srcConfig); err != nil {
		return err
	}

	s.BackupPath = util.RenamePath(s.Path, "backup")
	if err := os.Rename(s.Path, s.BackupPath); err != nil {
		time.Sleep(10 * time.Second)
		if err := os.Rename(s.Path, s.BackupPath); err != nil {
			return err
		}
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

func (d *Dapp) fromConfig() error {
	configFile := filepath.Join(d.Path, d.Controller.Configuration)
	read, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer read.Close()

	jsonMap := make(map[string]interface{})

	json.NewDecoder(read).Decode(&jsonMap)

	db, ok := jsonMap["DB"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("DB params not found")
	}
	conn, ok := db["Conn"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("Conn params not found")
	}

	if dbname, ok := conn[dbName]; ok {
		d.DBEngine.DB.DBName = dbname.(string)
	}

	if user, ok := conn[dbUser]; ok {
		d.DBEngine.DB.User = user.(string)
	}

	if pwd, ok := conn[dbPassword]; ok {
		d.DBEngine.DB.Password = pwd.(string)
	}

	if port, ok := conn[dbPort]; ok {
		d.DBEngine.DB.Port = port.(string)
	}

	if hostname, ok := jsonMap[torHostname]; ok {
		d.Tor.Hostname = hostname.(string)
	}

	if port, ok := jsonMap[torSocksListener]; ok {
		d.Tor.SocksPort, _ = strconv.Atoi(fmt.Sprintf("%v", port))
	}

	role, ok := jsonMap[role]
	if !ok {
		return fmt.Errorf("Role not found")
	}

	d.Role = role.(string)

	return nil
}

func (d *Dapp) setUIConfig() error {
	configFile := filepath.Join(d.Path, d.Gui.Configuration)
	read, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer read.Close()

	jsonMap := make(map[string]interface{})
	json.NewDecoder(read).Decode(&jsonMap)

	if err := setConfigurationValues(jsonMap, d.Gui.Settings); err != nil {
		return err
	}

	write, err := os.Create(configFile)
	if err != nil {
		return err
	}
	defer write.Close()

	return json.NewEncoder(write).Encode(jsonMap)
}

// ReadConfig reads dappctrl configuration file.
func (d *Dapp) ReadConfig() error {
	return d.fromConfig()
}
