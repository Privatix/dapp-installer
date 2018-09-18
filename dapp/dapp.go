package dapp

import (
	"path"

	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dapp-installer/windows"
)

// Dapp has a config for dapp core.
type Dapp struct {
	DownloadCtrl   string
	Controller     string
	Gui            string
	DownloadConfig string
	Configuration  string
	DownloadGui    string
	Installer      string
	InstallPath    string
	TempPath       string
	BackupPath     string
	Service        *windows.Service
	Shortcuts      bool
	UserRole       string
	Version        string
}

// NewConfig creates a default Dapp configuration.
func NewConfig() *Dapp {
	return &Dapp{}
}

// DownloadDappCtrl returns temporary downloaded path.
func (d *Dapp) DownloadDappCtrl() string {
	filePath := d.TempPath + path.Base(d.DownloadCtrl)
	if err := util.DownloadFile(filePath, d.DownloadCtrl); err != nil {
		return ""
	}

	return filePath
}
