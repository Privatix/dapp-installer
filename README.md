[![Go Report Card](http://goreportcard.com/badge/github.com/Privatix/dapp-installer)](https://goreportcard.com/report/github.com/Privatix/dapp-installer)
[![Maintainability](https://api.codeclimate.com/v1/badges/603af7ec449bf3ae153c/maintainability)](https://codeclimate.com/github/Privatix/dapp-installer/maintainability)
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

go get -d github.com/Privatix/dapp-installer

```

#### Windows platform

```bash
go get github.com/josephspurrier/goversioninfo/cmd/goversioninfo

go generate
``` 

# Usage

Simply run `dapp-installer <COMMAND>`

`dapp-installer` or `dapp-installer --help` will show usage and a list of commands:

```
Usage:
   dapp-installer [command] [flags]

 Available Commands:
   install     Install dapp core
   update      Update dapp core
   remove      Remove dapp core
 
 Flags:
   --help      Display help information
   --version   Display the current version of this CLI
   
 Use "dapp-installer [command] --help" for more information about a command.
 ```

# Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/Privatix/dappctrl/tags).

## Authors

* [ubozov](https://github.com/ubozov)

See also the list of [contributors](https://github.com/Privatix/dapp-installer/contributors) who participated in this project.

# License

This project is licensed under the **GPL-3.0 License** - see the [COPYING](COPYING) file for details.
