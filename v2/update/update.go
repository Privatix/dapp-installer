package update

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/privatix/dapp-installer/container"
	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dapp-installer/v2/flow"
	"github.com/privatix/dapp-installer/v2/metadata"
	"github.com/privatix/dapp-installer/v2/product"
	"github.com/privatix/dapp-installer/v2/service"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
)

var stepTimeout = 300 * time.Second

type updateContext struct {
	Role          string
	Path          string
	Source        string
	installed     *metadata.Installation
	updateVersion string
	path          appPath
}

type appPath struct {
	DappCtrl     dappctrlPath
	DappGUI      dappguiPath
	DB           dbPath
	Tor          torPath
	Logs         string
	Installation string
}

type dappctrlPath struct {
	EntryPoint string
	Config     string
}

type dappguiPath struct {
	Settings string
}

type dbPath struct {
	DataDir string
}

type torPath struct {
	HiddenService string
	Settings      string
}

type step struct {
	name string
	do   func(log.Logger, *updateContext) error
	undo func(log.Logger, *updateContext) error
}

func (s *step) Name() string { return s.name }

func (s *step) Do(l log.Logger, v interface{}) error {
	return validateAndRun(l, v, s.do)
}

func (s *step) Undo(l log.Logger, v interface{}) error {
	return validateAndRun(l, v, s.undo)
}

func validateAndRun(l log.Logger, v interface{}, f func(log.Logger, *updateContext) error) error {
	if f == nil {
		return nil
	}
	v2, ok := v.(*updateContext)
	if !ok {
		return fmt.Errorf("invlalid step context")
	}
	return f(l, v2)
}

// Run updates application.
func Run(logger log.Logger) error {
	v := &updateContext{
		installed: new(metadata.Installation),
		path: appPath{
			DappCtrl: dappctrlPath{
				EntryPoint: "dappctrl/dappctrl",
				Config:     "dappctrl/dappctrl.config.json",
			},
			DB: dbPath{
				DataDir: "pgsql/data/",
			},
			Tor: torPath{
				HiddenService: "tor/hidden_service/",
				Settings:      "tor/settings/",
			},
			Logs:         "log/",
			Installation: ".env.config.json",
		},
	}
	if runtime.GOOS == "linux" {
		v.path = appPath{
			DappCtrl: dappctrlPath{
				EntryPoint: "dappctrl/dappctrl",
				Config:     "dappctrl/dappctrl.config.json",
			},
			DappGUI: dappguiPath{
				Settings: "dappgui/resources/app/settings.json",
			},
			DB: dbPath{
				DataDir: "var/lib/postgresql",
			},
			Tor: torPath{
				HiddenService: "var/lib/tor/hidden_service/",
				Settings:      "etc/tor/",
			},
			Logs:         "log/",
			Installation: ".env.config.json",
		}
	} else if runtime.GOOS == "windows" {
		v.path = appPath{
			DappCtrl: dappctrlPath{
				EntryPoint: "dappctrl/dappctrl",
				Config:     "dappctrl/dappctrl.config.json",
			},
			DappGUI: dappguiPath{
				Settings: "dappgui/resources/app/settings.json",
			},
			DB: dbPath{
				DataDir: "pgsql/data/",
			},
			Tor: torPath{
				HiddenService: "tor/hidden_service/",
				Settings:      "tor/settings/",
			},
			Logs:         "log/",
			Installation: ".env.config.json",
		}
	} else if runtime.GOOS == "darwin" {
		v.path = appPath{
			DappCtrl: dappctrlPath{
				EntryPoint: "dappctrl/dappctrl",
				Config:     "dappctrl/dappctrl.config.json",
			},
			DappGUI: dappguiPath{
				Settings: "dappgui/dapp-gui.app/Contents/Resources/app/settings.json",
			},
			DB: dbPath{
				DataDir: "pgsql/data/",
			},
			Tor: torPath{
				HiddenService: "tor/hidden_service/",
				Settings:      "tor/settings/",
			},
			Logs:         "log/",
			Installation: ".env.config.json",
		}
	}
	updateFlow := flow.Flow{
		Name: "Update",
	}
	if runtime.GOOS == "linux" {
		updateFlow.Steps = []flow.Step{
			&step{
				name: "read config file",
				do:   readConfigFileAndArgs,
			},
			&step{
				name: "stop container",
				do:   stopLinuxContainer,
				undo: startLinuxContainer,
			},
			&step{
				name: "backup current installation",
				do:   backupCurrentInstallation,
				undo: restoreInstallationBackup,
			},
			&step{
				name: "extract new app files",
				do:   extractAppFilesForUpdate,
				undo: removeAppFilesForUpdate,
			},
			&step{
				name: "check updating to higher version",
				do:   setUpdateVersion,
			},
			&step{
				name: "copy some config values from installed to updating version",
				do:   mergeConfigs,
			},
			&step{
				name: "copy some gui values",
				do:   copyValuesForGUISettings,
			},
			&step{
				name: "copy database data to new app files",
				do:   copyDatabaseDataToNewAppFiles,
			},
			&step{
				name: "copy TOR configs to new app files",
				do:   copyTORConfigs,
			},
			&step{
				name: "start container",
				do:   startLinuxContainer,
				undo: stopLinuxContainer,
			},
			&step{
				name: "run database migrations within new app files",
				do:   updateDB,
			},
			&step{
				name: "update products within new app files",
				do:   updateProducts,
			},
			&step{
				name: "stop container",
				do:   stopLinuxContainer,
				undo: startLinuxContainer,
			},
			&step{
				name: "start container if agent",
				do:   startLinuxContainerIfAgent,
			},
		}
	} else {
		updateFlow.Steps = []flow.Step{
			&step{
				name: "read config file",
				do:   readConfigFileAndArgs,
			},
			&step{
				name: "read installed details",
				do:   readInstallationDetails,
			},
			&step{
				name: "stop TOR",
				do:   stopTor,
				undo: startTor,
			},
			&step{
				name: "stop all products",
				do:   stopAllProducts,
				undo: startAllProducts,
			},
			&step{
				name: "stop dappctrl",
				do:   stopDappCtrl,
				undo: startDappCtrl,
			},
			&step{
				name: "stop database",
				do:   stopDatabase,
				undo: startDatabase,
			},
			&step{
				name: "backup current installation",
				do:   backupCurrentInstallation,
				undo: restoreInstallationBackup,
			},
			&step{
				name: "extract new app files",
				do:   extractAppFilesForUpdate,
				undo: removeAppFilesForUpdate,
			},
			&step{
				name: "check updating to higher version",
				do:   setUpdateVersion,
			},
			&step{
				name: "copy some config values from installed to updating version",
				do:   mergeConfigs,
			},
			&step{
				name: "copy some gui values",
				do:   copyValuesForGUISettings,
			},
			&step{
				name: "copy database data to new app files",
				do:   copyDatabaseDataToNewAppFiles,
			},
			&step{
				name: "copy win dappctrl service files",
				do:   copyWinDappctrlServiceFiles,
			},
			&step{
				name: "run database to apply migrations",
				do:   startDatabase,
				undo: stopDatabase,
			},
			&step{
				name: "run database migrations within new app files and load new prod data",
				do:   updateDB,
			},
			&step{
				name: "stop database",
				do:   stopDatabaseIfClient,
			},
			&step{
				name: "copy TOR configs to new app files",
				do:   copyTORConfigs,
			},
			&step{
				name: "copy logs to new app files",
				do:   copyLogs,
			},
			&step{
				name: "update products within new app files",
				do:   updateProducts,
			},
			&step{
				name: "start TOR if agent",
				do:   startTorIfAgent,
			},
			&step{
				name: "start dappctrl if agent",
				do:   startDappCtrlIfAgent,
			},
			&step{
				name: "start all products if agent",
				do:   startAllProductsIfAgent,
			},
			&step{
				name: "save updated installation details wihtin new app files",
				do:   saveInstallationDetails,
			},
		}
	}
	return updateFlow.Run(logger, v)
}

func readConfigFileAndArgs(_ log.Logger, v *updateContext) error {
	conffile := flag.String("config", "dapp-installer.config.json", "dapp-installer configuration file")
	role := flag.String("role", "", "client | agent")
	workdir := flag.String("workdir", "", "app directory")

	flag.CommandLine.Parse(os.Args[2:])

	err := util.ReadJSON(*conffile, v)
	if err != nil {
		return fmt.Errorf("could not read configuration file: %v", err)
	}

	if *role != "" {
		v.Role = *role
	}

	if *workdir != "" {
		v.Path = *workdir
	}

	v.Path, err = filepath.Abs(v.Path)
	if err != nil {
		return fmt.Errorf("could not get absolute path for installation: %v", err)
	}

	return nil
}

func readInstallationDetails(logger log.Logger, v *updateContext) error {
	if err := util.ReadJSON(filepath.Join(v.Path, v.path.Installation), v.installed); err != nil {
		return fmt.Errorf("could not read installed details: %v", err)
	}
	logger.Info(fmt.Sprintf("Installatino: %+v", v.installed))
	return nil
}

func stopTor(logger log.Logger, v *updateContext) error {
	return stopService(logger, v.installed.Tor.Service)
}

func startTor(logger log.Logger, v *updateContext) error {
	return startService(logger, v.installed.Tor.Service)
}

func stopDappCtrl(logger log.Logger, v *updateContext) error {
	return stopService(logger, v.installed.Dapp.Service)
}

func startDappCtrl(logger log.Logger, v *updateContext) error {
	return startService(logger, v.installed.Dapp.Service)
}

func stopDatabase(logger log.Logger, v *updateContext) error {
	return stopService(logger, v.installed.DB.Service)
}

func startDatabase(logger log.Logger, v *updateContext) error {
	return startService(logger, v.installed.DB.Service)
}

func stopService(logger log.Logger, svc string) error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	logger = logger.Add("Uid", currentUser.Uid)
	ctx, cancel := context.WithTimeout(context.Background(), stepTimeout)
	defer cancel()
	if err := service.Stop(ctx, logger, svc, currentUser.Uid); err != nil {
		return fmt.Errorf("could not stop a service: %v", err)
	}
	return nil
}

func startService(logger log.Logger, svc string) error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	logger = logger.Add("Uid", currentUser.Uid)
	ctx, cancel := context.WithTimeout(context.Background(), stepTimeout)
	defer cancel()
	if err := service.Start(ctx, logger, svc, currentUser.Uid); err != nil {
		return fmt.Errorf("could not start a service: %v", err)
	}
	return nil
}

func currentInstallationBackupPath(v *updateContext) string {
	return v.Path + "." + "old"
}

func backupCurrentInstallation(_ log.Logger, v *updateContext) error {
	backupPath := currentInstallationBackupPath(v)
	// HACK: some files under sudo on darwin but execution is not.
	if runtime.GOOS == "darwin" {
		command := fmt.Sprintf("rm -rf %s && mv %s %s", backupPath, v.Path, backupPath)
		return util.ExecuteCommandOnDarwinAsAdmin(command)
	}
	// HACK: util.CopyDir doesn't work with some container files. Use command for now.
	if runtime.GOOS == "linux" {
		command := fmt.Sprintf("rm -rf %s && mv %s %s", backupPath, v.Path, backupPath)
		return util.ExecuteCommand("/bin/bash", "-c", command)
	}
	// Windows.
	if err := os.RemoveAll(currentInstallationBackupPath(v)); err != nil {
		return fmt.Errorf("could not prepare backup folder: %v", err)
	}
	if err := util.CopyDir(v.Path, currentInstallationBackupPath(v)); err != nil {
		return fmt.Errorf("could not backup current installation: %v", err)
	}
	if err := os.RemoveAll(v.Path); err != nil {
		return fmt.Errorf("could not clean up: %v", err)
	}
	return nil
}

func restoreInstallationBackup(_ log.Logger, v *updateContext) error {
	backupPath := currentInstallationBackupPath(v)
	// HACK: some files under sudo on darwin but execution is not.
	if runtime.GOOS == "darwin" {
		command := fmt.Sprintf("rm -rf %s && mv %s %s", v.Path, backupPath, v.Path)
		return util.ExecuteCommandOnDarwinAsAdmin(command)
	}
	// HACK: util.CopyDir doesn't work with some container files. Use command for now.
	if runtime.GOOS == "linux" {
		command := fmt.Sprintf("rm -rf %s && mv %s %s", v.Path, backupPath, v.Path)
		return util.ExecuteCommand("/bin/bash", "-c", command)
	}
	// Windows.
	if err := os.RemoveAll(v.Path); err != nil {
		return fmt.Errorf("could not prepare restore folder: %v", err)
	}
	if err := util.CopyDir(currentInstallationBackupPath(v), v.Path); err != nil {
		return fmt.Errorf("could not backup current installation: %v", err)
	}
	if err := os.RemoveAll(currentInstallationBackupPath(v)); err != nil {
		return fmt.Errorf("could not clean up: %v", err)
	}
	return nil
}

func stopAllProducts(logger log.Logger, v *updateContext) error {
	ctx, cancel := context.WithTimeout(context.Background(), stepTimeout)
	defer cancel()
	if err := product.StopAll(ctx, logger, v.Path, v.Role); err != nil {
		return fmt.Errorf("could not stop all products: %v", err)
	}
	return nil
}

func startAllProducts(logger log.Logger, v *updateContext) error {
	ctx, cancel := context.WithTimeout(context.Background(), stepTimeout)
	defer cancel()
	if err := product.StartAll(ctx, logger, v.Path, v.Role); err != nil {
		return fmt.Errorf("could not start all products: %v", err)
	}
	return nil
}

func extractAppFilesForUpdate(_ log.Logger, v *updateContext) error {
	if _, err := os.Stat(v.Path); os.IsNotExist(err) {
		os.MkdirAll(v.Path, util.FullPermission)
	}

	if err := util.Unzip(v.Source, v.Path); err != nil {
		return fmt.Errorf("could not to unzip source `%s`: %v", v.Source, err)
	}
	return nil
}

func removeAppFilesForUpdate(_ log.Logger, v *updateContext) error {
	if err := os.RemoveAll(v.Path); err != nil {
		return fmt.Errorf("could not remove update's app files: %v", err)
	}
	return nil
}

func setUpdateVersion(logger log.Logger, v *updateContext) error {
	f := filepath.Join(v.Path, v.path.DappCtrl.EntryPoint)
	version := util.DappCtrlVersion(f)
	if util.ParseVersion(version) <= util.ParseVersion(v.installed.Version) {
		logger.Warn(fmt.Sprintf("updating to: %s, installed: %s", version, v.installed.Version))
		return fmt.Errorf("update not required")
	}
	v.updateVersion = version
	return nil
}

func mergeConfigs(_ log.Logger, v *updateContext) error {
	copyItems := [][]string{
		[]string{"DB", "Conn", "port"},
		[]string{"Eth", "GethURL"},
		[]string{"Eth", "Contract"},
		[]string{"Report", "userid"},
		[]string{"FileLog", "Filename"},
		[]string{"FileLog", "Level"},
		[]string{"StaticPassword"},
		[]string{"TorHostname"},
		[]string{"TorSocksListener"},
		[]string{"PayAddress"},
		[]string{"PayServer"},
		[]string{"SOMCServer", "Addr"},
		[]string{"Sess", "Addr"},
		[]string{"UI", "Addr"},
		[]string{"Role"},
	}
	src := filepath.Join(currentInstallationBackupPath(v), v.path.DappCtrl.Config)
	dst := filepath.Join(v.Path, v.path.DappCtrl.Config)
	return util.UpdateConfig(copyItems, src, dst)
}

func copyValuesForGUISettings(_ log.Logger, v *updateContext) error {
	copyItems := [][]string{
		[]string{"bugsnag"},
		[]string{"lang"},
		[]string{"role"},
	}
	src := filepath.Join(currentInstallationBackupPath(v), v.path.DappGUI.Settings)
	dst := filepath.Join(v.Path, v.path.DappGUI.Settings)
	return util.UpdateConfig(copyItems, src, dst)
}

func copyDatabaseDataToNewAppFiles(_ log.Logger, v *updateContext) error {
	return copyDir(v, v.path.DB.DataDir)
}

func copyWinDappctrlServiceFiles(logger log.Logger, v *updateContext) error {
	if runtime.GOOS != "windows" {
		logger.Info("not windows, skipping")
		return nil
	}
	if err := copyFileFromDappctrlDir(v.installed.Dapp.Service+".config.json", v); err != nil {
		return err
	}

	if err := copyFileFromDappctrlDir(v.installed.Dapp.Service+".exe", v); err != nil {
		return err
	}
	return nil
}

func copyFileFromDappctrlDir(f string, v *updateContext) error {
	src := filepath.Join(currentInstallationBackupPath(v), "dappctrl", f)
	dst := filepath.Join(v.Path, "dappctrl", f)
	if err := util.CopyFile(src, dst); err != nil {
		return fmt.Errorf("could not copy `%s`: %v", f, err)
	}
	return nil
}

func updateDB(logger log.Logger, v *updateContext) error {
	dappctrl := filepath.Join(v.Path, v.path.DappCtrl.EntryPoint)
	dbconf := struct {
		DB *data.DBConfig
	}{
		DB: data.NewDBConfig(),
	}
	if err := util.ReadJSON(filepath.Join(v.Path, v.path.DappCtrl.Config), &dbconf); err != nil {
		return fmt.Errorf("could not read db config: %v", err)
	}

	// Ping DB.
	ctx, cancel := context.WithTimeout(context.Background(), stepTimeout)
	defer cancel()
	err := util.RetryTillSucceed(ctx, func() error {
		conn, err := sql.Open("postgres", dbconf.DB.ConnStr())
		if err == nil {
			defer conn.Close()
			err = conn.Ping()
		}
		return err
	})
	if err != nil {
		return fmt.Errorf("could not ping database: %v", err)
	}

	if err := runCommand(logger, dappctrl, "db-migrate", "-conn", dbconf.DB.ConnStr()); err != nil {
		return fmt.Errorf("could not run migrations: %v", err)
	}
	if err := runCommand(logger, dappctrl, "db-load-data", "-conn", dbconf.DB.ConnStr()); err != nil {
		return fmt.Errorf("could not load prod data: %v", err)
	}
	return nil
}

func runCommand(logger log.Logger, file string, args ...string) error {
	logger.Info("command: " + file + " " + strings.Join(args, " "))
	timeoutC := time.After(stepTimeout)
	var err error
	for {
		select {
		case <-timeoutC:
			return fmt.Errorf("timed out, last error: %v", err)
		default:
			if err = util.ExecuteCommand(file, args...); err == nil {
				return nil
			}
			logger.Warn(fmt.Sprintf("retry scheduled, got error: %v", err))
			time.Sleep(time.Second)
		}
	}
}

func stopDatabaseIfClient(logger log.Logger, v *updateContext) error {
	if v.Role == data.RoleClient {
		logger.Info("client, stop database")
		return stopDatabase(logger, v)
	}
	logger.Info("agent, won't stop database")
	return nil
}

func startTorIfAgent(logger log.Logger, v *updateContext) error {
	if v.Role == data.RoleAgent {
		logger.Info("agent, start tor")
		return startTor(logger, v)
	}
	logger.Info("client, won't start tor")
	return nil
}

func startDappCtrlIfAgent(logger log.Logger, v *updateContext) error {
	if v.Role == data.RoleAgent {
		logger.Info("agent, start dappctrl")
		return startDappCtrl(logger, v)
	}
	logger.Info("client, won't start dappctrl")
	return nil
}

func startAllProductsIfAgent(logger log.Logger, v *updateContext) error {
	if v.Role == data.RoleAgent {
		logger.Info("agent, start all products")
		return startAllProducts(logger, v)
	}
	logger.Info("client won't start any product")
	return nil
}

func copyTORConfigs(_ log.Logger, v *updateContext) error {
	if err := copyTORHiddenService(v); err != nil {
		return err
	}
	return copyTORSettings(v)
}

func copyTORHiddenService(v *updateContext) error {
	return copyDir(v, v.path.Tor.HiddenService)
}

func copyTORSettings(v *updateContext) error {
	return copyDir(v, v.path.Tor.Settings)
}

func copyLogs(_ log.Logger, v *updateContext) error {
	return copyDir(v, v.path.Logs)
}

func updateProducts(logger log.Logger, v *updateContext) error {
	ctx, cancel := context.WithTimeout(context.Background(), stepTimeout)
	defer cancel()
	backupPath := currentInstallationBackupPath(v)
	if err := product.UpdateAll(ctx, logger, backupPath, v.Path, v.Role); err != nil {
		return fmt.Errorf("could not update all products: %v", err)
	}
	return nil
}

func copyDir(v *updateContext, p string) error {
	backupPath := currentInstallationBackupPath(v)
	src := filepath.Join(backupPath, p)
	dst := filepath.Join(v.Path, p)
	if runtime.GOOS == "linux" {
		command := fmt.Sprintf("rm -rf %s && cp -rp %s %s", dst, src, dst)
		return util.ExecuteCommand("/bin/bash", "-c", command)
	}
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("could not prepare folder: %v", err)
	}
	if err := util.CopyDir(src, dst); err != nil {
		return fmt.Errorf("could not copy dir: %v", err)
	}
	return nil
}

func saveInstallationDetails(_ log.Logger, v *updateContext) error {
	tmp := *v.installed
	tmp.Version = v.updateVersion
	return util.WriteJSON(filepath.Join(v.Path, v.path.Installation), &tmp)
}

func startLinuxContainer(logger log.Logger, v *updateContext) error {
	c := getContainer(v.Role, v.Path)

	if err := c.Start(); err != nil {
		return fmt.Errorf("could not start container: %v", err)
	}
	return nil
}

func startLinuxContainerIfAgent(logger log.Logger, v *updateContext) error {
	if v.Role == data.RoleAgent {
		logger.Info("agent, starting container")
		return startLinuxContainer(logger, v)
	}

	logger.Info("client, won't start container")
	return nil
}

func stopLinuxContainer(_ log.Logger, v *updateContext) error {
	c := getContainer(v.Role, v.Path)

	if err := c.Stop(); err != nil {
		return fmt.Errorf("could not stop container: %v", err)
	}
	return nil
}

func getContainer(role, path string) *container.Container {
	c := container.NewContainer()

	c.Name = role
	c.Path = path
	return c
}
