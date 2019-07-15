// dapp-installer is a cli installer for dapp core.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/privatix/dapp-installer/dapp"
	"github.com/privatix/dapp-installer/flow"
	"github.com/privatix/dapp-installer/flows"
	"github.com/privatix/dapp-installer/v2/update"
	"github.com/privatix/dappctrl/util/log"
	"github.com/privatix/dappctrl/version"
)

// Values for versioning.
var (
	Commit  string
	Version string
)

func createLogger() (log.Logger, io.Closer) {
	failIfErr := func(err error) {
		if err != nil {
			panic(fmt.Sprintf("failed to create logger: %v", err))
		}
	}

	logConfig := &log.FileConfig{
		WriterConfig: log.NewWriterConfig(),
		Filename:     "dapp-installer-%Y-%m-%d.log",
		FileMode:     0644,
	}

	flog, closer, err := log.NewFileLogger(logConfig)
	failIfErr(err)

	logger := flog

	for _, value := range os.Args[1:] {
		if value == "-verbose" || value == "--verbose" {
			elog, err := log.NewStderrLogger(log.NewWriterConfig())
			failIfErr(err)
			logger = log.NewMultiLogger(elog, flog)
			break
		}
	}

	return logger, closer
}

func main() {
	logger, closer := createLogger()
	defer closer.Close()

	d := dapp.NewDapp()

	arg := ""
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}

	if arg == "help" {
		fmt.Println(flows.RootHelp)
		return
	}

	if arg == "update" {
		if err := update.Run(logger); err != nil {
			logger.Fatal(err.Error())
		}
		return
	}

	ok, err := flow.Execute(logger, arg, map[string]flow.Flow{
		"install":          flows.Install(),
		"install-products": flows.InstallProducts(),
		// "update":           flows.Update(),
		"update-products": flows.UpdateProducts(),
		"remove":          flows.Remove(),
		"remove-products": flows.RemoveProducts(),
		"status":          flows.Status(),
	}, d)

	if err != nil {
		object, _ := json.Marshal(d)
		logger = logger.Add("object", string(object))
		logger.Error(fmt.Sprintf("%v", err))
		if !d.Verbose {
			fmt.Println("error:", err)
			fmt.Println("object:", string(object))
		}
		os.Exit(2)
	}

	if !ok {
		v := flag.Bool("version", false, "Prints current dapp-installer version")

		flag.Parse()

		if *v {
			version.Print(true, Commit, Version)
			os.Exit(0)
		}

		fmt.Printf(flows.RootHelp)
	}
}
