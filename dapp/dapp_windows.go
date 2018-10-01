// +build windows

package dapp

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/privatix/dapp-installer/dbengine"
	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dapp-installer/windows"
	"github.com/privatix/dappctrl/util/log"
)

// NewConfig creates a default Dapp configuration.
func NewConfig() *Dapp {
	reg := &util.Registry{
		Install:   []util.Key{},
		Uninstall: []util.Key{},
	}

	return &Dapp{
		DBEngine: dbengine.NewConfig(),
		Registry: reg,
	}
}

type service struct {
	windows.Service
}

func (d *Dapp) createShortcut() {
	target := filepath.Join(d.InstallPath, d.Gui.EntryPoint)
	gui := path.Base(d.Gui.EntryPoint)
	extension := filepath.Ext(gui)
	linkName := gui[0 : len(gui)-len(extension)]
	link := filepath.Join(util.DesktopPath(),
		fmt.Sprintf("%s-%s.lnk", linkName, d.UserRole))

	windows.CreateShortcut(target, "Privatix Dapp", "", link)
}

func (d *Dapp) configurateController(logger log.Logger) error {
	if d.Gui.Shortcuts {
		d.createShortcut()
	}

	if err := d.modifyDappConfig(logger); err != nil {
		return err
	}

	_, installer := filepath.Split(os.Args[0])
	util.CopyFile(os.Args[0], filepath.Join(d.InstallPath, installer))

	logger.Info("create service wrapper")
	ctrl := d.Controller
	ctrlPath := filepath.Join(d.InstallPath,
		filepath.Dir(d.Controller.EntryPoint))

	hash := util.Hash(d.InstallPath)
	ctrl.Service.ID = fmt.Sprintf("dapp_%s_%s", d.UserRole, hash)
	ctrl.Service.Name = fmt.Sprintf("dapp %s %s", d.UserRole, hash)
	ctrl.Service.Description = fmt.Sprintf("dapp %s %s", d.UserRole, hash)
	ctrl.Service.GUID = filepath.Join(ctrlPath, ctrl.Service.ID)
	if err := ctrl.Service.CreateWrapper(ctrlPath); err != nil {
		logger.Warn("failed to create service wrapper:" + ctrl.Service.ID)
		return err
	}
	logger.Info("install service")
	if err := ctrl.Service.Install(); err != nil {
		logger.Warn("failed to install service:" + ctrl.Service.ID)
		return err
	}
	logger.Info("start service")
	if err := ctrl.Service.Start(); err != nil {
		// retry attempt to start service
		if err := ctrl.Service.Start(); err != nil {
			logger.Warn("failed to start service:" + ctrl.Service.ID)
			return err
		}
	}

	return nil
}

func (d *Dapp) removeFinalize(logger log.Logger) error {
	if d.Gui.Shortcuts {
		gui := path.Base(d.Gui.EntryPoint)
		extension := filepath.Ext(gui)
		linkName := gui[0 : len(gui)-len(extension)]
		link := filepath.Join(util.DesktopPath(),
			fmt.Sprintf("%s-%s.lnk", linkName, d.UserRole))
		os.Remove(link)
	}

	util.RemoveRegistryKey(d.UserRole)

	return nil
}

func (d *Dapp) installFinalize(logger log.Logger) error {
	db := d.DBEngine.DB
	shortcuts := strconv.FormatBool(d.Gui.Shortcuts)
	reg := d.Registry.(*util.Registry)
	reg.Install = append(reg.Install,
		util.Key{Name: "Shortcuts", Type: "string", Value: shortcuts},
		util.Key{Name: "BaseDirectory", Type: "string", Value: d.InstallPath},
		util.Key{Name: "Version", Type: "string", Value: d.Version},
		util.Key{Name: "ServiceID", Type: "string",
			Value: d.Controller.Service.GUID},
		util.Key{Name: "Controller", Type: "string",
			Value: d.Controller.EntryPoint},
		util.Key{Name: "Gui", Type: "string", Value: d.Gui.EntryPoint},
		util.Key{Name: "Database", Type: "string", Value: db.DBName},
		util.Key{Name: "Configuration", Type: "string",
			Value: d.Controller.Configuration},
	)

	current := fmt.Sprintf("%d%d%d", time.Now().Year(),
		time.Now().Month(), time.Now().Day())

	_, installer := filepath.Split(os.Args[0])
	uninstallCmd := fmt.Sprintf("%s remove -role %s",
		filepath.Join(d.InstallPath, installer), d.UserRole)
	size, err := util.DirSize(d.InstallPath)
	if err != nil {
		return err
	}
	reg.Uninstall = append(reg.Uninstall,
		util.Key{Name: "InstallLocation", Type: "string",
			Value: d.InstallPath},
		util.Key{Name: "InstallDate", Type: "string", Value: current},
		util.Key{Name: "DisplayVersion", Type: "string", Value: d.Version},
		util.Key{Name: "DisplayName", Type: "string",
			Value: "Privatix Dapp " + d.UserRole},
		util.Key{Name: "UninstallString", Type: "string",
			Value: uninstallCmd},
		util.Key{Name: "EstimatedSize", Type: "dword",
			Value: strconv.FormatInt(size, 10)},
	)
	return util.CreateRegistryKey(reg, d.UserRole)
}

func (d *Dapp) updateFinalize(logger log.Logger) error {
	current := fmt.Sprintf("%d%d%d", time.Now().Year(),
		time.Now().Month(), time.Now().Day())

	size, err := util.DirSize(d.InstallPath)
	if err != nil {
		return err
	}

	reg := &util.Registry{
		Install: []util.Key{
			util.Key{Name: "Version", Type: "string", Value: d.Version},
		},
		Uninstall: []util.Key{
			util.Key{Name: "InstallDate", Type: "string", Value: current},
			util.Key{Name: "DisplayVersion", Type: "string", Value: d.Version},
			util.Key{Name: "EstimatedSize", Type: "dword",
				Value: strconv.FormatInt(size, 10)},
		},
	}

	d.Registry = reg
	return util.CreateRegistryKey(reg, d.UserRole)
}

func (d *Dapp) removeRegistry() error {
	return util.RemoveRegistryKey(d.UserRole)
}

// Exists returns existing dapp in the host.
func Exists(role string, logger log.Logger) (*Dapp, bool) {
	maps, ok := util.ExistingDapp(role, logger)

	if !ok {
		return nil, false
	}

	shortcut, _ := strconv.ParseBool(maps["Shortcuts"])
	d := &Dapp{
		UserRole:    role,
		Version:     maps["Version"],
		InstallPath: maps["BaseDirectory"],
		Controller: &InstallerEntity{
			Configuration: maps["Configuration"],
			Service:       &service{windows.Service{GUID: maps["ServiceID"]}},
		},
		Gui: &InstallerEntity{
			EntryPoint: maps["Gui"],
			Shortcuts:  shortcut,
		},
	}
	return d, true
}
