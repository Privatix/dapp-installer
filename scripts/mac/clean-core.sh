#!/bin/sh

cd /Applications || exit 1

# client
launchctl stop io.privatix.ctrl_8af2c39def3354a568cccfbf48903a52d2a61a20
launchctl stop io.privatix.tor_8af2c39def3354a568cccfbf48903a52d2a61a20
launchctl stop io.privatix.db_8af2c39def3354a568cccfbf48903a52d2a61a20

launchctl remove io.privatix.ctrl_8af2c39def3354a568cccfbf48903a52d2a61a20
launchctl remove io.privatix.tor_8af2c39def3354a568cccfbf48903a52d2a61a20
launchctl remove io.privatix.db_8af2c39def3354a568cccfbf48903a52d2a61a20

rm -rf "$HOME/Library/LaunchAgents/io.privatix.ctrl_8af2c39def3354a568cccfbf48903a52d2a61a20.plist"
rm -rf "$HOME/Library/LaunchAgents/io.privatix.tor_8af2c39def3354a568cccfbf48903a52d2a61a20.plist"
rm -rf "$HOME/Library/LaunchAgents/io.privatix.db_8af2c39def3354a568cccfbf48903a52d2a61a20.plist"

rm -Rf ./Privatix/client

rm -Rf ./Privatix

rm -Rf './Privatix client'

# agent
launchctl stop io.privatix.ctrl_7136467c58b0b8421697bccc29e48826ba831936
launchctl stop io.privatix.tor_7136467c58b0b8421697bccc29e48826ba831936
launchctl stop io.privatix.db_7136467c58b0b8421697bccc29e48826ba831936

launchctl remove io.privatix.ctrl_7136467c58b0b8421697bccc29e48826ba831936
launchctl remove io.privatix.tor_7136467c58b0b8421697bccc29e48826ba831936
launchctl remove io.privatix.db_7136467c58b0b8421697bccc29e48826ba831936

rm -rf "$HOME/Library/LaunchAgents/io.privatix.ctrl_7136467c58b0b8421697bccc29e48826ba831936.plist"
rm -rf "$HOME/Library/LaunchAgents/io.privatix.tor_7136467c58b0b8421697bccc29e48826ba831936.plist"
rm -rf "$HOME/Library/LaunchAgents/io.privatix.db_7136467c58b0b8421697bccc29e48826ba831936.plist"

rm -Rf ./Privatix/agent

rm -Rf ./Privatix

rm -Rf './Privatix agent'

echo "core was removed"






