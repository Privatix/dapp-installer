package container

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/privatix/dapp-installer/statik"
	"github.com/privatix/dapp-installer/util"
)

// Container has a config for conatiner.
type Container struct {
	Name string
	Path string
}

// NewContainer creates a default Container configuration.
func NewContainer() *Container {
	return &Container{
		Name: "common",
		Path: "./common",
	}
}

func (c *Container) daemonName() string {
	return "systemd-nspawn@" + c.Name + ".service"
}

func (c *Container) daemonPath() string {
	return filepath.Join("/etc/systemd/system", c.daemonName())
}

// Install installs a container.
func (c *Container) Install() error {
	file, err := os.Create(c.daemonPath())
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := statik.ReadFile("/templates/container.tpl")
	if err != nil {
		return err
	}

	t, err := template.New("containerTemplate").Parse(string(data))
	if err != nil {
		return err
	}

	if err := t.Execute(file, &c); err != nil {
		return err
	}

	return util.ExecuteCommand("systemctl", "daemon-reload")
}

// EnableAndStart container nspawn service.
func (c *Container) EnableAndStart() error {
	// --now start a service
	return util.ExecuteCommand("systemctl", "enable", "--now", c.daemonName())
}

// DisableAndStop container nspawn service.
func (c *Container) DisableAndStop() error {
	// --now stops a service
	return util.ExecuteCommand("systemctl", "disable", "--now", c.daemonName())
}

// Start starts the container.
func (c *Container) Start() error {
	return util.ExecuteCommand("systemctl", "start", c.daemonName())
}

// Restart restarts the container.
func (c *Container) Restart() error {
	return util.ExecuteCommand("systemctl", "restart", c.daemonName())
}

// Stop stops the container.
func (c *Container) Stop() error {
	return util.ExecuteCommand("systemctl", "stop", c.daemonName())
}

// Remove removes the container.
func (c *Container) Remove() error {
	if c.IsActive() {
		return errors.New("can't remove active container")
	}

	if err := util.ExecuteCommand("systemctl", "disable",
		c.daemonName()); err != nil {
		return err
	}
	return os.Remove(c.daemonPath())
}

// IsActive returns the container active status.
func (c *Container) IsActive() bool {
	output, err := util.ExecuteCommandOutput("systemctl", "is-active",
		c.daemonName())
	if err != nil {
		return false
	}

	return strings.EqualFold("active", output)
}

// Update upgrades the container.
func (c *Container) Update(newPath, oldPath string, copies []string,
	merges []string) error {
	if err := os.Rename(c.Path, oldPath); err != nil {
		return fmt.Errorf("failed to backup container: %v", err)
	}

	// copies dirs
	for _, value := range copies {
		src := filepath.Join(oldPath, value)
		dst := filepath.Join(newPath, value, "..")
		if err := util.ExecuteCommand("cp", "-pRf", src, dst); err != nil {
			return fmt.Errorf("failed to copy data: %v", err)
		}
	}

	// merges configs
	for _, value := range merges {
		dst := filepath.Join(newPath, value)
		src := filepath.Join(oldPath, value)
		if err := util.MergeJSONFile(dst, src, "PSCAddrHex"); err != nil {
			return err
		}
	}

	return os.Rename(newPath, c.Path)
}
