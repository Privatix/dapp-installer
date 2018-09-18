// +build windows

package dapp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/privatix/dapp-installer/data"
	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dapp-installer/windows"
	"github.com/privatix/dappctrl/util/log"
)

// Install installs a dapp core.
func (d *Dapp) Install(db *data.DB, logger log.Logger, ok bool) error {
	if _, err := os.Stat(d.InstallPath); os.IsNotExist(err) {
		os.MkdirAll(d.InstallPath, 0644)
	}
	logger.Info("download and modify files")
	if err := downloadAndConfigurateFiles(d, db); err != nil {
		logger.Warn("ocurred error when downloaded files")
		return err
	}

	logger.Info("create service wrapper")
	d.Service.GUID = fmt.Sprintf("%s%s", d.InstallPath, d.Service.ID)
	if err := d.Service.CreateWrapper(d.InstallPath); err != nil {
		logger.Warn("ocurred error when create service wrapper:" +
			d.Service.ID)
		return err
	}
	logger.Info("install service")
	if err := d.Service.Install(); err != nil {
		logger.Warn("ocurred error when install service " + d.Service.ID)
		return err
	}
	logger.Info("start service")
	if err := d.Service.Start(); err != nil {
		logger.Warn("ocurred error when start service " + d.Service.ID)
		return err
	}
	return nil
}

// Remove removes installed dapp core.
func (d *Dapp) Remove() error {
	d.Service.Stop()
	d.Service.Remove()
	return os.RemoveAll(d.InstallPath)
}

func downloadAndConfigurateFiles(dapp *Dapp, db *data.DB) error {
	_, dapp.Controller = filepath.Split(dapp.DownloadCtrl)
	err := util.CopyFile(dapp.TempPath+dapp.Controller,
		dapp.InstallPath+dapp.Controller)
	if err != nil {
		return err
	}
	_, dapp.Gui = filepath.Split(dapp.DownloadGui)
	err = util.DownloadFile(dapp.InstallPath+dapp.Gui, dapp.DownloadGui)
	if err != nil {
		return err
	}
	if dapp.Shortcuts {
		dapp.createShortcut()
	}

	_, dapp.Installer = filepath.Split(os.Args[0])
	err = util.CopyFile(os.Args[0], dapp.InstallPath+dapp.Installer)
	if err != nil {
		return err
	}

	_, dapp.Configuration = filepath.Split(dapp.DownloadConfig)
	configFile := dapp.InstallPath + dapp.Configuration
	err = util.DownloadFile(configFile, dapp.DownloadConfig)
	if err != nil {
		return err
	}
	return modifyDappConfig(configFile, db, dapp.UserRole)
}

func modifyDappConfig(configFile string, dbConf *data.DB, role string) error {
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

	if _, ok := jsonMap["Role"]; ok {
		jsonMap["Role"] = role
	}

	write, err := os.Create(configFile)
	if err != nil {
		fmt.Println(err)
	}
	defer write.Close()

	return json.NewEncoder(write).Encode(jsonMap)
}

func (d *Dapp) createShortcut() {
	target := d.InstallPath + d.Gui
	extension := filepath.Ext(d.Gui)
	linkName := d.Gui[0 : len(d.Gui)-len(extension)]
	link := util.DesktopPath() + "\\" + linkName + ".lnk"

	windows.CreateShortcut(target, "Privatix Dapp", "", link)
}
