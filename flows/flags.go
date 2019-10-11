package flows

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/privatix/dapp-installer/dapp"
)

func processedInstallFlags(d *dapp.Dapp) error {
	return processedCommonFlags(d, installHelp)
}

func processedInstallProductFlags(d *dapp.Dapp) error {
	return processProductFlags(d, installProductHelp)
}

func processedRemoveProductFlags(d *dapp.Dapp) error {
	return processProductFlags(d, removeProductHelp)
}

func processedRemoveFlags(d *dapp.Dapp) error {
	return processedWorkFlags(d, removeHelp)
}

func processProductFlags(d *dapp.Dapp, help string) error {
	h := flag.Bool("help", false, "Display dapp-installer help")
	role := flag.String("role", "", "Dapp user role")
	path := flag.String("workdir", "", "Dapp install directory")
	product := flag.String("product", "", "Specific product")
	src := flag.String("source", "", "Dapp install source")
	uid := flag.String("uid", "", "installation user's UID")

	v := flag.Bool("verbose", false, "Display log to console output")

	flag.CommandLine.Parse(os.Args[2:])

	if *h {
		fmt.Println(help)
		os.Exit(0)
	}

	if err := validateUIDFlag(d, *uid); err != nil {
		return err
	}

	d.Verbose = *v

	if len(*role) > 0 {
		d.Role = *role
	} else {
		return errors.New("role is required")
	}

	if len(*path) > 0 {
		d.Path = *path
	} else {
		return errors.New("workdir is required")
	}

	if len(*src) > 0 {
		d.Source = *src
	} else if os.Args[1] == "install" {
		return errors.New("source is required")
	}

	if len(*product) > 0 {
		d.Product = *product
	}

	d.DBEngine.Autostart = d.Role == "agent"

	return nil
}

func processedCommonFlags(d *dapp.Dapp, help string) error {
	h := flag.Bool("help", false, "Display dapp-installer help")
	role := flag.String("role", "", "Dapp user role")
	path := flag.String("workdir", "", "Dapp install directory")
	src := flag.String("source", "", "Dapp install source")
	core := flag.Bool("core", false, "Install only dapp core")
	product := flag.String("product", "", "Specific product")
	sendremote := flag.Bool("sendremote", false, "Send error reports")
	torHSD := flag.String("torhsd", "", "Tor hidden service directory")
	torSocks := flag.String("torsocks", "", "Tor socks port number")
	uid := flag.String("uid", "", "installation user's UID")

	v := flag.Bool("verbose", false, "Display log to console output")

	flag.CommandLine.Parse(os.Args[2:])

	if err := validateUIDFlag(d, *uid); err != nil {
		return err
	}

	if *h {
		fmt.Println(help)
		os.Exit(0)
	}

	d.Verbose = *v

	if *torHSD == "" {
		return errors.New("torhsd is required")
	}
	d.Tor.HiddenServiceDir = *torHSD
	if *torSocks == "" {
		return errors.New("torsocks is required")
	}
	socksN, err := strconv.ParseInt(*torSocks, 10, 64)
	if err != nil {
		return fmt.Errorf("could not parse socks port number: %v", err)
	}
	d.Tor.SocksPort = int(socksN)

	if len(*role) > 0 {
		d.Role = *role
	} else {
		return errors.New("role is required")
	}

	if len(*path) > 0 {
		d.Path = *path
	} else {
		return errors.New("workdir is required")
	}

	if len(*src) > 0 {
		d.Source = *src
	} else {
		return errors.New("source is required")
	}

	if len(*product) > 0 {
		d.Product = *product
	}

	if sendremote != nil && *sendremote == true {
		d.SendRemote = *sendremote
	}

	d.OnlyCore = *core

	d.DBEngine.Autostart = d.Role == "agent"

	return nil
}

func processedStatusFlags(d *dapp.Dapp) error {
	return processedWorkFlags(d, statusHelp)
}

func processedWorkFlags(d *dapp.Dapp, help string) error {
	h := flag.Bool("help", false, "Display dapp-installer help")
	p := flag.String("workdir", "", "Dapp install directory")
	role := flag.String("role", "", "Dapp user role")
	uid := flag.String("uid", "", "installation user's UID")

	v := flag.Bool("verbose", false, "Display log to console output")

	flag.CommandLine.Parse(os.Args[2:])

	if err := validateUIDFlag(d, *uid); err != nil {
		return err
	}

	if *h {
		fmt.Println(help)
		os.Exit(0)
	}

	if len(*p) == 0 {
		*p = filepath.Dir(os.Args[0])
	}
	d.Path = *p
	d.Role = *role
	d.Verbose = *v
	return nil
}

func validateUIDFlag(d *dapp.Dapp, uid string) error {
	if runtime.GOOS == "darwin" {
		if uid == "" {
			return errors.New("UID argument is required")
		}
		u, err := user.LookupId(uid)
		if err != nil {
			return fmt.Errorf("could not find user with uid: %v: %v", uid, err)
		}
		d.Username = u.Username
		d.UID = uid
	}
	return nil
}
