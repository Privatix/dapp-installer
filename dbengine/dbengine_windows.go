// +build windows

package dbengine

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"
	"strings"

	"github.com/privatix/dapp-installer/util"
	"github.com/privatix/dappctrl/util/log"
)

// Install installs a DB engine.
func (engine *DBEngine) Install(installPath string, logger log.Logger) error {
	fileName := installPath + path.Base(engine.Download)
	if err := util.DownloadFile(fileName, engine.Download); err != nil {
		logger.Warn("ocurred error when downloded file: " + engine.Download)
		return err
	}
	logger.Info("file successfully downloaded")

	// install db engine
	ch := make(chan bool)
	defer close(ch)
	go interactiveWorker("Installation DB Engine", ch)

	if err := exec.Command(fileName,
		engine.generateInstallParams()...).Run(); err != nil {
		ch <- true
		fmt.Printf("\r%s\n", "Ocurred error when install DB Engine")
		logger.Warn("ocurred error when install dbengine")
		return err
	}
	logger.Info("dbengine successfully installed")

	ch <- true
	fmt.Printf("\r%s\n", "DB Engine successfully installed")

	for _, c := range engine.Copy {
		fmt.Println(c.From, c.To)
		fileName := path.Base(c.From)
		if err := util.DownloadFile(c.To+"\\"+fileName, c.From); err != nil {
			logger.Warn("ocurred error when downloded file from " + c.From)
			return err
		}
	}

	// start db engine service
	if err := startService(engine.ServiceName); err != nil {
		logger.Warn("ocurred error when start dbengine service")
		return err
	}

	logger.Info("dbengine service successfully started")
	return nil
}

func startService(service string) error {
	checkServiceCmd := exec.Command("sc", "queryex", service)

	var checkServiceStdOut bytes.Buffer
	checkServiceCmd.Stdout = &checkServiceStdOut

	if err := checkServiceCmd.Run(); err != nil {
		return err
	}

	// service is running
	if strings.Contains(checkServiceStdOut.String(), "RUNNING") {
		return nil
	}

	// trying start service
	return exec.Command("net", "start", service).Run()
}
