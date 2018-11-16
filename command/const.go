package command

// status constants
const (
	status  = "\t%20s:\t%s\n"
	success = "[OK]"
	failed  = "[FAILED]"
)

const envFile = ".env"

const rootHelp = `
Usage:
  dapp-installer [command] [flags]

Available Commands:
	install		Install dapp core
	update		Update dapp core
	remove		Remove dapp core
	status		Display dapp installation info

Flags:
	--help		Display help information
	--version	Display the current version of this CLI

Use "dapp-installer [command] --help" for more information about a command.
`

const updateHelp = `
Usage:
	dapp-installer update [flags]

Flags:
	--config	Configuration file
	--workdir	Dapp install directory
	--source	Dapp install source
	--help		Display help information
`

const removeHelp = `
Usage:
	dapp-installer remove [flags]

Flags:
	--help		Display help information
	--workdir	Dapp install directory
`

const installHelp = `
Usage:
	dapp-installer install [flags]

Flags:
	--config	Configuration file
	--role		Dapp user role
	--workdir	Dapp install directory
	--source	Dapp install source
	--help		Display help information
`

const statusHelp = `
Usage:
	dapp-installer status [flags]

Flags:
	--help		Display help information
	--workdir	Dapp install directory
`
