package container

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

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

	t, err := template.New("containerTemplate").Parse(containerTemplate)
	if err != nil {
		return err
	}

	if err := t.Execute(file, &c); err != nil {
		return err
	}

	if err := util.ExecuteCommand("systemctl", "daemon-reload"); err != nil {
		return err
	}

	return util.ExecuteCommand("systemctl", "enable", c.daemonName())
}

// Start starts the container.
func (c *Container) Start() error {
	return util.ExecuteCommand("systemctl", "start", c.daemonName())
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
		if err := util.MergeJSONFile(dst, src); err != nil {
			return err
		}
	}

	return os.Rename(newPath, c.Path)
}

var containerTemplate = `#  This file is part of systemd.
#
#  systemd is free software; you can redistribute it and/or modify it
#  under the terms of the GNU Lesser General Public License as published by
#  the Free Software Foundation; either version 2.1 of the License, or
#  (at your option) any later version.

[Unit]
Description=Container %i
Documentation=man:systemd-nspawn(1)
PartOf=machines.target
Before=machines.target
After=network.target


[Service]
Restart=always
ExecStartPre={{.Path}}/dappctrl/pre-start.sh
ExecStart=/usr/bin/systemd-nspawn --quiet --boot --keep-unit --machine={{.Name}} --link-journal=try-guest --directory={{.Path}} --bind=/lib/modules --bind=/etc/localtime:/etc/localtime --bind=/etc/resolv.conf:/etc/resolv.conf --settings=override --capability=CAP_NET_ADMIN --capability=CAP_NET_BIND_SERVICE --capability=CAP_SYS_ADMIN --capability=CAP_MAC_ADMIN --capability=CAP_SYS_MODULE
ExecStopPost={{.Path}}/dappctrl/post-stop.sh
KillMode=mixed
Type=notify
RestartForceExitStatus=133
SuccessExitStatus=133
# Enforce a strict device policy, similar to the one nspawn configures when it
# allocates its own scope unit. Make sure to keep these policies in sync if you
# change them!
DevicePolicy=closed
DeviceAllow=/dev/net/tun rwm
DeviceAllow=char-pts rw

# nspawn itself needs access to /dev/loop-control and /dev/loop, to implement
# the --image= option. Add these here, too.
DeviceAllow=/dev/loop-control rw
DeviceAllow=block-loop rw
DeviceAllow=block-blkext rw

# nspawn can set up LUKS encrypted loopback files, in which case it needs
# access to /dev/mapper/control and the block devices /dev/mapper/*.
DeviceAllow=/dev/mapper/control rw
DeviceAllow=block-device-mapper rw


[Install]
WantedBy=machines.target
`
