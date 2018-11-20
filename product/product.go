package product

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	dapputil "github.com/privatix/dappctrl/util"
)

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
func Install(role, path string) error {
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
			// todo import
			// err := writeVariable(envPath, productImport, "true")
			// if err != nil {
			// 	return err
			// }
		}

		if !installed {
			if coreDapp, err := install(role, productPath); err != nil {
				if coreDapp {
					return err
				}

				continue
			}

			err := writeVariable(envPath, productInstall, "true")
			if err != nil {
				return err
			}
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
			if err := update(role, productPath); err != nil {
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
		_, _, installed := getParameters(productPath)

		if installed {
			if err := remove(role, productPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func install(role, path string) (bool, error) {
	return run(role, path, "install")

	// configPath, err := findFile(path, configFile)
	// if err != nil {
	// 	return true, err
	// }

	// if len(configPath) == 0 {
	// 	return true, fmt.Errorf("%s not found in %s", configFile, path)
	// }

	// p := &product{}
	// if err := dapputil.ReadJSONFile(configPath, &p); err != nil {
	// 	return true, err
	// }

	// for _, v := range p.Install {
	// 	if err := v.execute(path); err != nil {
	// 		return p.CoreDapp, err
	// 	}
	// }

	// return p.CoreDapp, nil
}

func update(role, path string) error {
	_, err := run(role, path, "update")

	return err
}

func remove(role, path string) error {
	_, err := run(role, path, "remove")

	return err

	// configPath, err := findFile(path, configFile)
	// if err != nil {
	// 	return err
	// }

	// if len(configPath) == 0 {
	// 	return fmt.Errorf("%s not found in %s", configFile, path)
	// }

	// p := &product{}
	// if err := dapputil.ReadJSONFile(configPath, &p); err != nil {
	// 	return err
	// }

	// for _, v := range p.Remove {
	// 	if err := v.execute(path); err != nil {
	// 		return err
	// 	}
	// }

	// return nil
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
	if v.Admin && runtime.GOOS == "darwin" {
		txt := `with prompt "Privatix wants to make changes"`
		evelate := "with administrator privileges"
		script := fmt.Sprintf(`'do shell script "%s" %s %s'`,
			v.Command, txt, evelate)

		cmd := exec.Command("osascript", "-e", script)
		return cmd.Run()
	}

	cmds := strings.Split(strings.Replace(v.Command, "..",
		path, -1), " ")
	file := filepath.Join(path, cmds[0])
	cmd := exec.Command(file, cmds[1:]...)

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	if err := cmd.Run(); err != nil {
		fmt.Println("out:", outb.String())
		fmt.Println("err:", errb.String())
		fmt.Println(file, v.Command, "error", err.Error())
		return err
	}
	return nil
}
