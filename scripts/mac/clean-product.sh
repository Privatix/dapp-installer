#!/bin/sh

cd /Applications || exit 1

# client
sudo launchctl stop io.privatix.dappvpn_bb37df9964ce4f9d9748ddb395e116ac47a64b0a
sudo launchctl remove io.privatix.dappvpn_bb37df9964ce4f9d9748ddb395e116ac47a64b0a

sudo rm -rf /Library/LaunchDaemons/io.privatix.dappvpn_bb37df9964ce4f9d9748ddb395e116ac47a64b0a.plist

sudo rm -Rf ./Privatix/client/product

# agent
sudo launchctl stop io.privatix.dappvpn_e63203ea9155f944c2f861257963f9a03c94724e
sudo launchctl stop io.privatix.openvpn_e63203ea9155f944c2f861257963f9a03c94724e
sudo launchctl stop io.privatix.nat_e63203ea9155f944c2f861257963f9a03c94724e
sudo launchctl remove io.privatix.dappvpn_e63203ea9155f944c2f861257963f9a03c94724e
sudo launchctl remove io.privatix.openvpn_e63203ea9155f944c2f861257963f9a03c94724e
sudo launchctl remove io.privatix.nat_e63203ea9155f944c2f861257963f9a03c94724e

sudo rm -rf /Library/LaunchDaemons/io.privatix.dappvpn_e63203ea9155f944c2f861257963f9a03c94724e.plist
sudo rm -rf /Library/LaunchDaemons/io.privatix.openvpn_e63203ea9155f944c2f861257963f9a03c94724e.plist
sudo rm -rf /Library/LaunchDaemons/io.privatix.nat_e63203ea9155f944c2f861257963f9a03c94724e.plist

sudo rm -Rf ./Privatix/agent/product

echo "product was removed"
