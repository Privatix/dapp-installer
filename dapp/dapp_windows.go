// +build windows

package dapp

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/privatix/dapp-installer/data"
	"github.com/privatix/dapp-installer/util"
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
	err := util.DownloadFile(dapp.InstallPath+dapp.Controller,
		dapp.DownloadCtrl)
	if err != nil {
		return err
	}
	configFile := dapp.InstallPath + path.Base(dapp.DownloadConfig)
	err = util.DownloadFile(configFile, dapp.DownloadConfig)
	if err != nil {
		return err
	}
	return modifyDappConfig(configFile, db)
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
