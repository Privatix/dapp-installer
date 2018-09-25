package util

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	// MinAvailableDiskSize - available min 500MB
	MinAvailableDiskSize uint64 = 500 * 1024 * 1024
	// MinMemorySize  - min RAM 2 GB
	MinMemorySize uint64 = 2 * 1000 * 1024 * 1024
)

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
	cmd := exec.Command(filename, "-version")
	var output bytes.Buffer
	cmd.Stdout = &output
	if err := cmd.Run(); err != nil {
		return ""
	}

	version := output.String()
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
		return fmt.Sprintf(`%s\temporary\`, volume)
	}

	path := fmt.Sprintf(`%s\%x\`, volume, b)

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
		return err
	}
	defer srcfd.Close()
	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()
	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

// DirSize returns total dir size.
func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path,
		func(_ string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				size += info.Size()
			}
			return err
		})
	return size, err
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

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src string, dest string) ([]string, error) {
	var filenames []string
	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		defer rc.Close()
		fpath := filepath.Join(dest, f.Name)

		if !strings.HasPrefix(fpath,
			filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0644)
		} else {
			if err := extractFile(fpath, rc, f); err != nil {
				return filenames, err
			}
		}
	}
	return filenames, nil
}

func extractFile(fpath string, rc io.ReadCloser, f *zip.File) error {
	err := os.MkdirAll(filepath.Dir(fpath), 0644)
	if err != nil {
		return err
	}
	outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		f.Mode())
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, rc)

	return err
}

// CopyDir copies data from source directory to desctination directory.
func CopyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}
	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
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
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// InteractiveWorker shows display text for imitate interactive mode.
func InteractiveWorker(s string, quit chan bool) {
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

// FreePort returns available free port number.
func FreePort(port string) (string, error) {
	p, err := strconv.Atoi(port)
	if err != nil {
		return "", err
	}

	for i := p; i < 65535; i++ {
		ln, err := net.Listen("tcp", "localhost:"+strconv.Itoa(i))

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
