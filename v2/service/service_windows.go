package service

import (
	"github.com/privatix/dapp-installer/util"
)

func startService(service, _ string) error {
	return util.ExecuteCommand("sc", "start", service)
}

func stopService(service, _ string) error {
	return util.ExecuteCommand("sc", "stop", service)
}
