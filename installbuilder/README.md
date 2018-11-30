# Getting Started

This instruction is for compiling a graphical installer.

## Prerequisites

Install [Bitrock InstallBuilder](https://installbuilder.bitrock.com/) 

## Installation steps

Clone the `dapp-installer` repository using git:

```bash
git clone https://github.com/privatix/dapp-installer.git
cd dapp-installer
git checkout master
```

Copy binary packages to folders:
 
Mac OS:
```
installbuilder/mac-dapp-installer/app.zip
installbuilder/mac-dapp-installer/dapp-installer
installbuilder/mac-dapp-installer/dapp-installer.config.json
```

Windows: 
```
installbuilder/mac-dapp-installer/app.zip
installbuilder/mac-dapp-installer/dapp-installer.exe
installbuilder/mac-dapp-installer/dapp-installer.config.json
```

Run build Bitrock Installer:

```
cd installbuilder/project
[Path to InstallBuilder]/bin/builder build Privatix.xml windows  
```

