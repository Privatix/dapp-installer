// +build windows

package dapp

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dapp-installer/windows"
	"github.com/privatix/dappctrl/util/log"
)

type service struct {
	windows.Service
}

func newService() *service {
	return &service{
		windows.Service{
			Command:   "dappctrl",
			Args:      []string{"-config", "dappctrl.config.json"},
			AutoStart: true,
		},
	}
}

func (d *Dapp) createShortcut() {
	target := filepath.Join(d.InstallPath, d.Gui.EntryPoint)
	gui := path.Base(d.Gui.EntryPoint)
	extension := filepath.Ext(gui)
	linkName := gui[0 : len(gui)-len(extension)]
	link := filepath.Join(util.DesktopPath(),
		fmt.Sprintf("%s-%s.lnk", linkName, d.UserRole))

	if len(extension) == 0 {
		target += ".exe"
	}

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
	ctrlPath := filepath.Join(d.InstallPath, filepath.Dir(ctrl.EntryPoint))

	hash := d.controllerHash()
	ctrl.Service.ID = hash
	ctrl.Service.Name = hash
	ctrl.Service.Description = fmt.Sprintf("dapp controller %s", hash)
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

	return nil
}

// TODO (ubozov) this algorithm should be reviewed.
// Installs run-time components (the Visual C++ Redistributable Packages
// for VS 2013) that are required to run postgresql database engine.
func (d *Dapp) prepareToInstall(logger log.Logger) error {
	fileName := filepath.Join(d.InstallPath, "util/vcredist_x64.exe")
	args := []string{"/install", "/quiet", "/norestart"}
	return util.ExecuteCommand(fileName, args)
}

func copyServiceWrapper(d, s *Dapp) {
	hash := s.controllerHash()
	scvExe := "dappctrl/" + hash + ".exe"
	scvConfig := "dappctrl/" + hash + ".config.json"

	util.CopyFile(filepath.Join(s.InstallPath, scvExe),
		filepath.Join(d.InstallPath, scvExe))
	util.CopyFile(filepath.Join(s.InstallPath, scvConfig),
		filepath.Join(d.InstallPath, scvConfig))
}
