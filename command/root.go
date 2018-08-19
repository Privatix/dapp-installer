package command

import (
	"fmt"
	"log"

	"github.com/privatix/dapp-installer/util"
	"github.com/spf13/cobra"
)

// Config has a configuration params for installer
type Config struct {
	DB  string
	Log string
}

func initialize() *cobra.Command {
	var configFile string // config file location
	var showVers bool     // whether to print version info or not
	config := newConfig()

	readConfig := func(ccmd *cobra.Command, args []string) error {
		// if --config is passed, attempt to parse the config file
		if configFile != "" {
			err := util.ReadJSONFile(configFile, &config)
			if err != nil {
				log.Fatalln(err)
				return fmt.Errorf(
					"Failed to read config file - %s",
					err,
				)
			}
		}

		return nil
	}

	preRun := func(ccmd *cobra.Command, args []string) error {
		// if --version is passed print the version info
		if showVers {
			fmt.Printf("dapp-installer %s \n", util.Version())
			return fmt.Errorf("")
		}
		return nil
	}

	rootCmd := &cobra.Command{
		Use:               "dapp-installer",
		Short:             "dapp-installer - installer for dapp core",
		Long:              "dapp-installer - installer for dapp core",
		SilenceErrors:     true,
		SilenceUsage:      true,
		PersistentPreRunE: readConfig,
		PreRunE:           preRun,
		RunE:              run,
	}

	// cli-only flags
	rootCmd.Flags().BoolVarP(&showVers, "version", "v", false,
		"Display the current version of this CLI")

	installCmd := createInstallCmd()

	installCmd.Flags().StringVarP(&configFile, "config", "c", "",
		"Path to config file (with extension)")
	rootCmd.AddCommand(installCmd)

	updateCmd := createUpdateCmd()
	updateCmd.Flags().StringVarP(&configFile, "config", "c", "",
		"Path to config file (with extension)")
	rootCmd.AddCommand(updateCmd)

	removeCmd := createRemoveCmd()
	removeCmd.Flags().StringVarP(&configFile, "config", "c", "",
		"Path to config file (with extension)")
	rootCmd.AddCommand(removeCmd)

	return rootCmd
}

// NewConfig is create new config object
func newConfig() *Config {
	return &Config{
		DB:  "db config",
		Log: "logger",
	}
}

func run(ccmd *cobra.Command, args []string) error {
	ccmd.HelpFunc()(ccmd, args)
	return nil
}

// Execute command
func Execute() {
	rootCmd := initialize()

	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
		return
	}
}
