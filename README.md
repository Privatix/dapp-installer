[![Go report](https://goreportcard.com/badge/github.com/Privatix/dapp-installer)](https://goreportcard.com/badge/github.com/Privatix/dapp-installer)
[![Maintainability](https://api.codeclimate.com/v1/badges/af4e29689d76d8ccf974/maintainability)](https://codeclimate.com/github/Privatix/dapp-installer/maintainability)
[![GoDoc](https://godoc.org/github.com/Privatix/dapp-installer?status.svg)](https://godoc.org/github.com/Privatix/dapp-installer)

# Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

## Prerequisites

Install [Golang](https://golang.org/doc/install). Make sure that `$HOME/go/bin` (default `GOPATH`) is added to system path `$PATH`.

## Installation steps

Clone the `dapp-installer` repository using git:

```bash
git clone https://github.com/Privatix/dapp-installer.git
cd dapp-installer
git checkout master
```

Build `dapp-installer` package:

```bash
git clone https://github.com/Privatix/dapp-installer.git
cd dapp-installer

go get github.com/spf13/cobra
```

# Usage

Simply run `dapp-installer <COMMAND>`

`dapp-installer` or `dapp-installer -h` will show usage and a list of commands:

```
dapp-installer - installer for dapp core

Usage:
  dapp-installer [flags]
  dapp-installer [command]

Available Commands:
  help        Help about any command
  install     Install dapp core to host
  remove      Remove dapp core to host
  update      Update dapp core to host

Flags:
  -h, --help      help for dapp-installer
  -v, --version   Display the current version of this CLI

Use "dapp-installer [command] --help" for more information about a command.
```

