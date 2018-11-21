package product

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/reform.v1"

	dappdata "github.com/privatix/dappctrl/data"
	dapputil "github.com/privatix/dappctrl/util"

	"github.com/privatix/dapp-installer/util"
)

//go:generate go generate ../vendor/github.com/privatix/dappctrl/data/schema.go

type product struct {
	CoreDapp bool
	Install  []command
	Update   []command
	Remove   []command
}

type command struct {
	Admin   bool
	Command string
}

// Install installs the products.
func Install(role, path, conn string) error {
	path = filepath.Join(path, productDir)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	for _, f := range files {
		if !f.IsDir() {
			continue
		}

		productPath := filepath.Join(path, f.Name())
		envPath, imported, installed := getParameters(productPath)

		if !imported {
			templatePath := filepath.Join(productPath, "template")
			if err := addProduct(templatePath, conn); err != nil {
				return err
			}
			err := writeVariable(envPath, productImport, true)
			if err != nil {
				return err
			}

			src := agentAdapterConfig
			if role == "client" {
				src = clientAdapterConfig
			}
			src = filepath.Join(templatePath, src)
			configPath := filepath.Join(productPath, "config")
			dest := filepath.Join(configPath, adapterConfig)
			if err := util.CopyFile(src, dest); err != nil {
				return err
			}
		}

		if installed {
			continue
		}

		coreDapp, err := run(role, productPath, "install")
		if err != nil {
			if coreDapp {
				return err
			}
			continue
		}

		err = writeVariable(envPath, productInstall, true)
		if err != nil {
			return err
		}
	}
	return nil
}

// Update updates the products.
func Update(role, path string) error {
	path = filepath.Join(path, productDir)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		productPath := filepath.Join(path, f.Name())
		_, _, installed := getParameters(productPath)

		if installed {
			_, err := run(role, productPath, "update")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Remove removes the products.
func Remove(role, path string) error {
	path = filepath.Join(path, productDir)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		productPath := filepath.Join(path, f.Name())
		envPath, _, installed := getParameters(productPath)

		if !installed {
			continue
		}

		if _, err := run(role, productPath, "remove"); err != nil {
			return err
		}

		err := writeVariable(envPath, productInstall, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func run(role, path, cmd string) (bool, error) {
	configPath, err := findFile(path, configFile)
	if err != nil {
		return true, err
	}

	if len(configPath) == 0 {
		return true, fmt.Errorf("%s not found in %s", configFile, path)
	}

	p := &product{}
	if err := dapputil.ReadJSONFile(configPath, &p); err != nil {
		return true, err
	}

	var cmds []command
	switch cmd {
	case "install":
		cmds = p.Install
	case "update":
		cmds = p.Update
	case "remove":
		cmds = p.Remove
	default:
		return true, fmt.Errorf("unknown command %s", cmd)

	}

	for _, v := range cmds {
		if err := v.execute(path); err != nil {
			return p.CoreDapp, err
		}
	}

	return p.CoreDapp, nil
}

func (v command) execute(path string) error {
	n := strings.Split(strings.Replace(v.Command, "..",
		path, -1), " ")
	file := filepath.Join(path, n[0])

	cmd := exec.Command(file, n[1:]...)

	if v.Admin && runtime.GOOS == "darwin" {
		txt := `with prompt "Privatix wants to make changes"`
		evelate := "with administrator privileges"
		command := fmt.Sprintf("%s %s", file, strings.Join(n[1:], " "))
		script := fmt.Sprintf(`do shell script "sudo %s" %s %s`,
			command, txt, evelate)

		cmd = exec.Command("osascript", "-e", script)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute %s: %v", v.Command, err)
	}
	return nil
}

func addProduct(path, conn string) error {
	db, err := dappdata.NewDBFromConnStr(conn)
	if err != nil {
		return err
	}
	defer dappdata.CloseDB(db)

	err = db.InTransaction(func(t *reform.TX) error {
		return processor(path, true, t)
	})

	return err
}
