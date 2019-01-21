// +build linux

package container

import (
	"errors"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	return t.Execute(file, &c)
}

// Start starts the container.
func (c *Container) Start() error {
	if err := exec.Command("systemctl", "enable",
		c.daemonName()).Run(); err != nil {
		return err
	}

	cmd := exec.Command("systemctl", "start", c.daemonName())
	return cmd.Run()
}

// Stop stops the container.
func (c *Container) Stop() error {
	cmd := exec.Command("systemctl", "stop", c.daemonName())
	return cmd.Run()
}

// Remove removes the container.
func (c *Container) Remove() error {
	if c.IsActive() {
		return errors.New("can't remove active container")
	}

	return os.Remove(c.daemonPath())
}

// IsActive returns the container active status.
func (c *Container) IsActive() bool {
	cmd := exec.Command("systemctl", "is-active", c.daemonName())
	output, err := cmd.Output()

	if err != nil {
		return false
	}

	return strings.EqualFold("active", string(output))
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
ExecStart=/usr/bin/systemd-nspawn --quiet --boot --keep-unit --machine={{.Name}} --link-journal=try-guest --directory={{.Path}} --bind=/lib/modules --bind=/etc/localtime:/etc/localtime --bind=/etc/resolv.conf:/etc/resolv.conf --settings=override --capability=CAP_NET_ADMIN --capability=CAP_NET_BIND_SERVICE --capability=CAP_SYS_ADMIN --capability=CAP_MAC_ADMIN --capability=CAP_SYS_MODULE
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
WantedBy=machine.target
`
