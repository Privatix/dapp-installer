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

## Settings

The `forceUpdate` parameter is used to control the update mode.
Set `1` to enable the forced update mode:
```
        <booleanParameter>
            <name>forceUpdate</name>
            <description></description>
            <explanation></explanation>
            <value>1</value>
            <default>0</default>
            <ask>0</ask>
        </booleanParameter>
```
or set `0` to disable the forced update mode:
```
        <booleanParameter>
            <name>forceUpdate</name>
            <description></description>
            <explanation></explanation>
            <value>0</value>
            <default>0</default>
            <ask>0</ask>
        </booleanParameter>
```
