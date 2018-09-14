// +build windows

package dapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
	_, dapp.Gui = filepath.Split(dapp.DownloadGui)
	err = util.DownloadFile(dapp.InstallPath+dapp.Gui, dapp.DownloadGui)
	if err != nil {
		return err
	}
	if dapp.Shortcuts {
		dapp.createShortcut()
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

func (d *Dapp) createShortcut() {
	fileName := path.Base(d.DownloadGui)
	path := d.InstallPath
	target := path + fileName
	extension := filepath.Ext(fileName)
	linkName := fileName[0 : len(fileName)-len(extension)]
	destination := util.DesktopPath()
	arguments := ""

	var scriptTxt bytes.Buffer
	scriptTxt.WriteString(fmt.Sprintf(`
	option explicit

	sub CreateShortCut()
		dim objShell, objLink
		set objShell = CreateObject("WScript.Shell")
		set objLink = objShell.CreateShortcut("%s\%s.lnk")
		objLink.Arguments = "%s"
		objLink.Description = "Privatix Dapp"
		objLink.TargetPath = "%s"
		objLink.WindowStyle = 1
		objLink.WorkingDirectory = "%s"
		objLink.Save
	end sub

	call CreateShortCut()`, destination, linkName, arguments, target, path))

	scriptFile := fmt.Sprintf("lnkTo%s.vbs", linkName)
	ioutil.WriteFile(scriptFile, scriptTxt.Bytes(), 0777)
	cmd := exec.Command("wscript", scriptFile)
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
	os.Remove(scriptFile)
	return
}
