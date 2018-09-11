package util

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/privatix/dapp-installer/data"
)

const (
	// MinAvailableDiskSize - available min 500MB
	MinAvailableDiskSize uint64 = 500 * 1024 * 1024
	// MinMemorySize  - min RAM 2 GB
	MinMemorySize uint64 = 2 * 1000 * 1024 * 1024
)

// WriteCounter type is used for download process.
type WriteCounter struct {
	Processed uint64
	Total     uint64
}

// DBEngine has a db engine configuration.
type DBEngine struct {
	Download    string
	ServiceName string
	DataDir     string
	InstallDir  string
	Copy        []Copy
	DB          *data.DB
}

// Copy has a file copies parameters.
type Copy struct {
	From string
	To   string
}

// DappCtrlConfig has a config for dappctrl.
type DappCtrlConfig struct {
	Download string
	File     string
	Version  string
	Service  *serviceConfig
}

type serviceConfig struct {
	ID          string
	Name        string
	Description string
	Command     string
	Args        []string
	AutoStart   bool
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Processed += uint64(n)
	wc.printProgress()
	return n, nil
}

func (wc WriteCounter) printProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rDownloading... %s from %s",
		humanateBytes(wc.Processed), humanateBytes(wc.Total))
}

func downloadFile(filePath, url string) error {
	out, err := os.Create(filePath + ".tmp")
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	length, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return err
	}

	counter := &WriteCounter{Total: uint64(length)}

	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}

	fmt.Print("\n")

	out.Close()

	return os.Rename(filePath+".tmp", filePath)
}

func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}

func humanateBytes(s uint64) string {
	sizes := []string{"B", "kB", "MB", "GB", "TB", "PB", "EB"}
	var base float64 = 1024
	if s < 10 {
		return fmt.Sprintf("%d B", s)
	}
	e := math.Floor(logn(float64(s), base))
	suffix := sizes[int(e)]
	val := math.Floor(float64(s)/math.Pow(base, e)*10+0.5) / 10
	f := "%.0f %s"
	if val < 10 {
		f = "%.1f %s"
	}

	return fmt.Sprintf(f, val, suffix)
}

func generateDBEngineInstallParams(dbConf *DBEngine) []string {
	args := []string{"--mode", "unattended", "--unattendedmodeui", "none"}

	if len(dbConf.ServiceName) > 0 {
		args = append(args, "--servicename", dbConf.ServiceName)
	}
	if len(dbConf.DB.User) > 0 {
		args = append(args, "--superaccount", dbConf.DB.User)
	}
	if len(dbConf.DB.Password) > 0 {
		args = append(args, "--superpassword", dbConf.DB.Password)
	}
	if len(dbConf.InstallDir) > 0 {
		args = append(args, "--prefix", dbConf.InstallDir)
	}
	if len(dbConf.DataDir) > 0 {
		args = append(args, "--datadir", dbConf.DataDir)
	}
	return args
}

func interactiveWorker(s string, quit chan bool) {
	i := 0
	for {
		select {
		case <-quit:
			return
		default:
			i++
			fmt.Printf("\r%s", strings.Repeat(" ", len(s)+15))
			fmt.Printf("\r%s%s", s, strings.Repeat(".", i))
			if i >= 10 {
				i = 0
			}
			time.Sleep(time.Millisecond * 250)
		}
	}
}

// GetDappCtrlVersion returns dappctrl version.
func GetDappCtrlVersion(filename string) string {
	cmd := exec.Command(filename, "-version")
	var output bytes.Buffer
	cmd.Stdout = &output
	if err := cmd.Run(); err != nil {
		return ""
	}
	return output.String()
}
