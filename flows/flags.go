package flows

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	dapputil "github.com/privatix/dappctrl/util"

	"github.com/privatix/dapp-installer/dapp"
)

func processedInstallFlags(d *dapp.Dapp) error {
	return processedCommonFlags(d, installHelp)
}

func processedUpdateFlags(d *dapp.Dapp) error {
	return processedCommonFlags(d, updateHelp)
}

func processedInstallProductFlags(d *dapp.Dapp) error {
	return processedCommonFlags(d, installProductHelp)
}

func processedUpdateProductFlags(d *dapp.Dapp) error {
	return processedCommonFlags(d, updateProductHelp)
}

func processedRemoveProductFlags(d *dapp.Dapp) error {
	return processedCommonFlags(d, removeProductHelp)
}

func processedRemoveFlags(d *dapp.Dapp) error {
	return processedWorkFlags(d, removeHelp)
}

func processedCommonFlags(d *dapp.Dapp, help string) error {
	h := flag.Bool("help", false, "Display dapp-installer help")
	config := flag.String("config", "", "Configuration file")
	role := flag.String("role", "", "Dapp user role")
	path := flag.String("workdir", "", "Dapp install directory")
	src := flag.String("source", "", "Dapp install source")
	core := flag.Bool("core", false, "Install only dapp core")
	product := flag.String("product", "", "Specific product")

	v := flag.Bool("verbose", false, "Display log to console output")

	flag.CommandLine.Parse(os.Args[2:])

	if *h {
		fmt.Println(help)
		os.Exit(0)
	}

	d.Verbose = *v

	if len(*config) > 0 {
		if err := dapputil.ReadJSONFile(*config, &d); err != nil {
			return err
		}
	}

	if len(*role) > 0 {
		d.Role = *role
	}

	if len(*path) > 0 {
		d.Path = *path
	}

	if len(*src) > 0 {
		d.Source = *src
	}

	if len(*product) > 0 {
		d.Product = *product
	}

	d.OnlyCore = *core

	return nil
}

func processedStatusFlags(d *dapp.Dapp) error {
	return processedWorkFlags(d, statusHelp)
}

func processedWorkFlags(d *dapp.Dapp, help string) error {
	h := flag.Bool("help", false, "Display dapp-installer help")
	p := flag.String("workdir", "", "Dapp install directory")
	config := flag.String("config", "", "Configuration file")

	v := flag.Bool("verbose", false, "Display log to console output")

	flag.CommandLine.Parse(os.Args[2:])

	if *h {
		fmt.Println(help)
		os.Exit(0)
	}

	if len(*config) > 0 {
		if err := dapputil.ReadJSONFile(*config, &d); err != nil {
			return err
		}
	}

	if len(*p) == 0 {
		*p = filepath.Dir(os.Args[0])
	}
	d.Path = *p
	d.Verbose = *v
	return nil
}
