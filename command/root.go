package command

import (
	"fmt"

	"github.com/privatix/dapp-installer/util"
	"github.com/spf13/cobra"
)

type config struct {
	DB  string
	Log string
}

var (
	conf = newConfig()

	fconfig  string // config file location
	showVers bool   // whether to print version info or not

	version string = "0.0.0.0" //todo

	// RootCmd ...
	RootCmd = &cobra.Command{
		Use:               "dapp-installer",
		Short:             "dapp-installer - installer for dapp core",
		Long:              "dapp-installer - installer for dapp core",
		SilenceErrors:     true,
		SilenceUsage:      true,
		PersistentPreRunE: readConfig,
		PreRunE:           preRun,
		RunE:              run,
	}
)

func newConfig() *config {
	return &config{
		DB:  "db config",
		Log: "logger",
	}
}

func readConfig(ccmd *cobra.Command, args []string) error {
	// if --config is passed, attempt to parse the config file
	if fconfig != "" {
		if err := util.ReadJSONFile(fconfig, &conf); err != nil {
			return fmt.Errorf("Failed to read config file - %s", err)
		}
	}

	return nil
}

func preRun(ccmd *cobra.Command, args []string) error {
	// if --version is passed print the version info
	if showVers {
		fmt.Printf("dapp-installer %s \n", version)
		return fmt.Errorf("")
	}
	return nil
}

func init() {
	// cli-only flags
	RootCmd.Flags().BoolVarP(&showVers, "version", "v", false, "Display the current version of this CLI")

	// commands flags
	installCmd.Flags().StringVarP(&fconfig, "config", "c", "", "Path to config file (with extension)")
	updateCmd.Flags().StringVarP(&fconfig, "config", "c", "", "Path to config file (with extension)")
	removeCmd.Flags().StringVarP(&fconfig, "config", "c", "", "Path to config file (with extension)")

	// commands
	RootCmd.AddCommand(installCmd)
	RootCmd.AddCommand(updateCmd)
	RootCmd.AddCommand(removeCmd)
}

func run(ccmd *cobra.Command, args []string) error {
	ccmd.HelpFunc()(ccmd, args)
	return nil
}
