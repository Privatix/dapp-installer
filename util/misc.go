package util

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	ps "github.com/keybase/go-ps"
)

const (
	// MinAvailableDiskSize - available min 500MB
	MinAvailableDiskSize uint64 = 500 * 1024 * 1024
	// MinMemorySize  - min RAM 2 GB
	MinMemorySize uint64 = 2 * 1000 * 1024 * 1024
	// FullPermission - 0777
	FullPermission os.FileMode = 0777
)

// Addr type contains address, host and port parameters.
type Addr struct {
	Address string
	Host    string
	Port    string
}

// WriteCounter type is used for download process.
type WriteCounter struct {
	Label        string
	Processed    uint64
	Total        uint64
	OutputLenght int
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Processed += uint64(n)
	wc.printProgress()
	return n, nil
}

func (wc *WriteCounter) printProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", wc.OutputLenght+1))
	output := fmt.Sprintf("\rDownloading %s ... %s from %s",
		wc.Label, humanateBytes(wc.Processed), humanateBytes(wc.Total))
	wc.OutputLenght = len(output)
	fmt.Printf(output)
}

// DownloadFile downloads the file.
func DownloadFile(filePath, url string) error {
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

	counter := &WriteCounter{Label: path.Base(url), Total: uint64(length)}

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

// DappCtrlVersion returns dappctrl version.
func DappCtrlVersion(filename string) string {
	version, err := ExecuteCommandOutput(filename, "-version")
	if err != nil {
		return ""
	}
	if i := strings.Index(version, " "); i > 0 {
		version = version[:i]
	}

	if strings.Contains(version, "undefined") {
		return "0.0.0"
	}
	return version
}

// TempPath creates temporary directory.
func TempPath(volume string) string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return filepath.Join(volume, "temporary")
	}

	if len(volume) == 0 {
		volume = "/tmp"
	}

	path := filepath.Join(volume, fmt.Sprintf("%x", b))

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0644)
	}

	return path
}

// CopyFile copies file.
func CopyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo
	if srcfd, err = os.Open(src); err != nil {
		return fmt.Errorf("could not open source file: %v", err)
	}
	defer srcfd.Close()
	if dstfd, err = os.Create(dst); err != nil {
		return fmt.Errorf("could not create destination file: %v", err)
	}
	defer dstfd.Close()
	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return fmt.Errorf("could not copy file: %v", err)
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return fmt.Errorf("could not read file stats: %v", err)
	}
	if err = os.Chmod(dst, srcinfo.Mode()); err != nil {
		return fmt.Errorf("could not change mode of a file: %v", err)
	}
	return nil
}

// ParseVersion returns version number in int64 format.
func ParseVersion(s string) int64 {
	strList := strings.Split(s, ".")
	length := len(strList)
	for i := 0; i < (4 - length); i++ {
		strList = append(strList, "0")
	}

	format := fmt.Sprintf("%%s%%0%ds", 4)
	v := ""
	for _, value := range strList {
		v = fmt.Sprintf(format, v, value)
	}

	result, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0
	}
	return result
}

// CopyDir copies data from source directory to desctination directory.
func CopyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo
	if srcinfo, err = os.Stat(src); err != nil {
		return fmt.Errorf("could not read dir status: %v", err)
	}
	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return fmt.Errorf("could not make dir: %v", err)
	}
	if fds, err = ioutil.ReadDir(src); err != nil {
		return fmt.Errorf("could not read dir: %v", err)
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())
		if fd.IsDir() {
			err = CopyDir(srcfp, dstfp)
		} else {
			err = CopyFile(srcfp, dstfp)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Hash returns string hash.
func Hash(s string) string {
	h := sha1.New()
	h.Write([]byte(strings.ToLower(s)))
	return hex.EncodeToString(h.Sum(nil))
}

// FreePort returns available free port number.
func FreePort(host, port string) (string, error) {
	p, err := strconv.Atoi(port)
	if err != nil {
		return "", err
	}

	for i := p; i < 65535; i++ {
		ln, err := net.Listen("tcp", host+":"+strconv.Itoa(i))

		if err != nil {
			continue
		}

		if err := ln.Close(); err != nil {
			continue
		}
		port = strconv.Itoa(i)
		break
	}

	return port, nil
}

// RenamePath changes folder name and returns it
func RenamePath(path, folder string) string {
	dir, _ := filepath.Split(path)
	return filepath.Join(dir, folder)
}

// IsURL checks path to url.
func IsURL(path string) bool {
	pattern := `^(https?:\/\/)`

	ok, _ := regexp.MatchString(pattern, path)
	return ok
}

// MatchAddr returns matches addr params from text.
func MatchAddr(str string) []Addr {
	pattern := `(?m)"Addr"\s*:\s*"\s*((localhost|127\.0\.0\.1|0\.0\.0\.0)\s*:\s*(\d{1,5}))\s*"`

	var addrs []Addr
	re := regexp.MustCompile(pattern)
	for _, match := range re.FindAllStringSubmatch(str, -1) {
		addrs = append(addrs, Addr{
			Address: match[1],
			Host:    match[2],
			Port:    match[3],
		})
	}
	return addrs
}

// KillProcess kills all processes at dir path.
func KillProcess(dir string) error {
	processes, err := ps.Processes()
	if err != nil {
		return err
	}
	execFile := filepath.Base(os.Args[0])

	for _, v := range processes {
		path, _ := v.Path()
		if path == "" || strings.EqualFold(execFile, filepath.Base(path)) {
			continue
		}
		processPath := filepath.ToSlash(strings.ToLower(path))
		if strings.HasPrefix(processPath, strings.ToLower(dir)) {
			proc, err := os.FindProcess(v.Pid())
			if err != nil {
				return err
			}

			if err := proc.Kill(); err != nil {
				return err
			}
		}
	}
	return nil
}

// SelfRemove removes itself execute file.
func SelfRemove(dir string) error {
	if runtime.GOOS != "windows" {
		return nil
	}

	path, err := filepath.Abs(os.Args[0])
	if err != nil {
		return err
	}
	exePath := filepath.ToSlash(strings.ToLower(path))

	if strings.HasPrefix(exePath, strings.ToLower(dir)) {
		cmd := fmt.Sprintf(`ping localhost -n 3 > nul & del %s`,
			filepath.Base(path))
		return exec.Command("cmd", "/c", cmd).Start()
	}
	return nil
}

// TimeOutInSec returns time duration in seconds.
func TimeOutInSec(timeout uint64) time.Duration {
	return time.Duration(timeout) * time.Second
}

// RetryTillSucceed tries execute func till succeed or returns timeout error.
func RetryTillSucceed(ctx context.Context, f func() error) (err error) {
	t := time.NewTicker(200 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			err = f()
			if err == nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// ReadJSON reads json file.
func ReadJSON(location string, v interface{}) error {
	data, err := ioutil.ReadFile(location)
	if err != nil {
		return fmt.Errorf("could not read `%s`: %v", location, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("could not unmarshall file content: %v", err)
	}
	return err
}

// WriteJSON reads json file.
func WriteJSON(location string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("could not encode to json: %v", err)
	}
	var mode os.FileMode = 0644
	fi, err := os.Stat(location)
	if err == nil {
		mode = fi.Mode()
	}
	if !os.IsNotExist(err) && err != nil {
		return fmt.Errorf("could not get `%s` mode: %v", location, err)
	}
	err = ioutil.WriteFile(location, data, mode)
	if err != nil {
		return fmt.Errorf("could not write `%s`: %v", location, err)
	}
	return nil
}

// UpdateConfig updates json config file from given source.
func UpdateConfig(copyItems [][]string, src, dst string) error {
	srcConfigMap := make(map[string]interface{})
	if err := ReadJSON(src, &srcConfigMap); err != nil {
		return fmt.Errorf("could not read installed dappctrl config at `%s`, got: %v", src, err)
	}

	dstConfigMap := make(map[string]interface{})
	if err := ReadJSON(dst, &dstConfigMap); err != nil {
		return fmt.Errorf("could not read new dappctrl config at `%s`, got: %v", dst, err)
	}
	for _, copyItem := range copyItems {
		var val, to interface{}
		var key string
		val = srcConfigMap
		to = dstConfigMap
		for _, subitem := range copyItem {
			if key != "" {
				tmp, ok := to.(map[string]interface{})[key]
				if !ok {
					return fmt.Errorf("could not get map by key `%s`", key)
				}
				to = tmp
			}
			key = subitem
			tmp, ok := val.(map[string]interface{})[subitem]
			if !ok {
				return fmt.Errorf("could not get map by key `%s`", key)
			}
			val = tmp
		}

		tmp, ok := to.(map[string]interface{})
		if !ok {
			return fmt.Errorf("could not get map by key `%s`", key)
		}
		tmp[key] = val
	}

	if err := WriteJSON(dst, &dstConfigMap); err != nil {
		return fmt.Errorf("could not write updated config: %v", err)
	}
	return nil
}

// ChownToUID changes owner of a path to user with given uid.
func ChownToUID(path string, installUID string) error {
	u, err := user.LookupId(installUID)
	if err != nil {
		return err
	}
	uid, err := strconv.ParseInt(u.Uid, 10, 64)
	if err != nil {
		return err
	}
	gid, err := strconv.ParseInt(u.Gid, 10, 64)
	if err != nil {
		return err
	}
	return os.Chown(path, int(uid), int(gid))
}
