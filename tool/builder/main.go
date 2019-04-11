package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/privatix/dappctrl/util"
)

type copyFile struct {
	From string
	// To path should comply https://github.com/Privatix/privatix/blob/master/doc/service_plug-in.md
	To string
}

type configuration struct {
	// PackageFilesDir files here should be placed according to
	// https://github.com/Privatix/privatix/blob/master/doc/service_plug-in.md
	PackageFilesDir string
	CopyFiles       []copyFile
}

const (
	help     = "builder -config=config.json -in=in_directory -out=out_directory"
	pathPerm = 0755
	filePerm = 0644
)

var (
	conffile        string
	sourceDirectory string
	outputDirectory string
	// https://github.com/Privatix/privatix/blob/master/doc/service_plug-in.md
	packageDirectoryMap = map[string]bool{
		"bin":      true,
		"config":   true,
		"data":     true,
		"log":      true,
		"template": true,
	}
	// https://github.com/Privatix/privatix/blob/master/doc/service_plug-in.md
	packageDirectoriesSlice = []string{"bin", "config", "data", "log", "template"}
)

func main() {
	flag.StringVar(&conffile, "config", "", "Configuration of package folder builder")
	flag.StringVar(&sourceDirectory, "in", "", "files location directory")
	flag.StringVar(&outputDirectory, "out", "", "output package directory")

	flag.Parse()

	if conffile == "" || sourceDirectory == "" || outputDirectory == "" {
		fmt.Println(conffile, sourceDirectory, outputDirectory)
		fmt.Println(help)
		return
	}

	conf := &configuration{}

	err := util.ReadJSONFile(conffile, &conf)
	if err != nil {
		panic(err)
	}

	validateFiles(conf)
	dirCopy(filepath.Join(sourceDirectory, conf.PackageFilesDir), outputDirectory)
	for _, item := range conf.CopyFiles {
		fileCopy(filepath.Join(sourceDirectory, item.From), filepath.Join(outputDirectory, item.To))
	}
	createMissingPackageDirs()
}

func createMissingPackageDirs() {
	contents, err := ioutil.ReadDir(outputDirectory)
	must(err)
	created := make(map[string]bool)
	for _, content := range contents {
		if !content.IsDir() {
			panic(fmt.Errorf("unexpected file in out directory: %s", content.Name()))
		}
		created[content.Name()] = true
	}

	for _, packageDir := range packageDirectoriesSlice {
		if !created[packageDir] {
			err := os.Mkdir(filepath.Join(outputDirectory, packageDir), pathPerm)
			must(err)
		}
	}
}

func validateFiles(conf *configuration) {
	validatePackageFiles(conf.PackageFilesDir)
	validateCopyFiles(conf.CopyFiles)
}

func validatePackageFiles(dir string) {
	contents, err := ioutil.ReadDir(filepath.Join(sourceDirectory, dir))
	must(err)

	for _, content := range contents {
		if content.IsDir() && !packageDirectoryMap[content.Name()] {
			panic(fmt.Sprintf("directory '%s' is not a package directory", content.Name()))
		}
	}
}

func validateCopyFiles(copyFiles []copyFile) {
	for _, item := range copyFiles {
		if !allowedDestination(item.To) {
			must(fmt.Errorf("unknown destination folder"))
		}
	}
}

func allowedDestination(dst string) bool {
	for _, dirname := range packageDirectoriesSlice {
		if strings.HasPrefix(dst, dirname) {
			return true
		}
	}
	return false
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func copy(src, dst string, info os.FileInfo) error {
	if info.IsDir() {
		return dirCopy(src, dst)
	}
	return fileCopy(src, dst)
}

func fileCopy(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), pathPerm); err != nil {
		return err
	}

	f, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, filePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	_, err = io.Copy(f, s)
	return err
}

func dirCopy(srcDir, dstDir string) error {
	if err := os.MkdirAll(dstDir, pathPerm); err != nil {
		return err
	}

	contents, err := ioutil.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, content := range contents {
		var base string
		if content.IsDir() {
			base = filepath.Base(content.Name())

		} else {
			base = content.Name()
		}

		cs := filepath.Join(srcDir, base)
		cd := filepath.Join(dstDir, base)

		if err := copy(cs, cd, content); err != nil {
			return err
		}
	}
	return nil
}
