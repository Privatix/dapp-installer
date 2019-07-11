package flows

// status constants
const (
	status  = "\t%20s:\t%s\n"
	success = "[OK]"
	failed  = "[FAILED]"
)

const envFile = ".env.config.json"
const dappCtrlDaemonOnLinux = "lib/systemd/system/dappctrl.service"

// Help messages.
const (
	RootHelp = `
Usage:
  dapp-installer [command] [flags]

Available Commands:
	install			Install dapp core and products
	update			Update dapp core and products
	remove			Remove dapp core and products
	install-products	Install products
	update-products		Update products
	remove-products		Remove products
	status			Display dapp installation info

Flags:
	--help			Display help information
	--version		Display the current version of this CLI
	--verbose		Display log to console log

Use "dapp-installer [command] --help" for more information about a command.
`
	updateHelp = `
Usage:
	dapp-installer update [flags]

Flags:
	--config	Configuration file
	--workdir	Dapp install directory
	--source	Dapp install source
	--verbose	Display log to console log
	--help		Display help information
`
	removeHelp = `
Usage:
	dapp-installer remove [flags]

Flags:
	--help		Display help information
	--workdir	Dapp install directory
	--verbose	Display log to console log
	--role			Dapp user role
`

	installHelp = `
Usage:
	dapp-installer install [flags]

Flags:
	--config		Configuration file
	--core			Install only dapp core
	--role			Dapp user role
	--workdir		Dapp install directory
	--source		Dapp install source
	--verbose		Display log to console log
	--help			Display help information
	--sendremote 	Whether to allow sending error reports or not
`

	statusHelp = `
Usage:
	dapp-installer status [flags]

Flags:
	--help		Display help information
	--workdir	Dapp install directory
	--verbose	Display log to console log
	--role			Dapp user role
`

	installProductHelp = `
Usage:
	dapp-installer install-products [flags]

Flags:
	--role		Dapp user role
	--workdir	Dapp install directory
	--product	Specific product
	--verbose	Display log to console log
	--help		Display help information
`

	updateProductHelp = `
Usage:
	dapp-installer update-products [flags]

Flags:
	--role		Dapp user role
	--workdir	Dapp install directory
	--product	Specific product
	--source	Path to new products
	--verbose	Display log to console log
	--help		Display help information
`

	removeProductHelp = `
Usage:
	dapp-installer remove-products [flags]

Flags:
	--role		Dapp user role
	--workdir	Dapp install directory
	--product	Specific product
	--verbose	Display log to console log
	--help		Display help information
`
)
