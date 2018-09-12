package dapp

import (
	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dapp-installer/windows"
)

// Dapp has a config for dapp core.
type Dapp struct {
	DownloadCtrl   string
	Controller     string
	DownloadConfig string
	DownloadGui    string
	InstallPath    string
	BackupPath     string
	Service        *windows.Service
	UserRole       string
	Version        string
}

// NewConfig creates a default Dapp configuration.
func NewConfig() *Dapp {
	return &Dapp{}
}

// DownloadDappCtrl returns temporary downloaded path.
func (d *Dapp) DownloadDappCtrl(path string) string {
	tempDappCtrl, err := util.TemporaryDownload(path, d.DownloadCtrl)
	if err != nil {
		return ""
	}

	return tempDappCtrl
}
