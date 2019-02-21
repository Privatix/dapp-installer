#  This file is part of systemd.
#
#  systemd is free software; you can redistribute it and/or modify it
#  under the terms of the GNU Lesser General Public License as published by
#  the Free Software Foundation; either version 2.1 of the License, or
#  (at your option) any later version.

[Unit]
Description=Container %i
Documentation=man:systemd-nspawn(1)
PartOf=machines.target
Before=machines.target
After=network.target
StartLimitIntervalSec=0

[Service]
Restart=always
ExecStartPre={{.Path}}/dappctrl/pre-start.sh
ExecStart=/usr/bin/systemd-nspawn --quiet --boot --keep-unit --machine={{.Name}} --link-journal=try-guest --directory={{.Path}} --bind=/lib/modules --bind=/etc/localtime:/etc/localtime --bind=/etc/resolv.conf:/etc/resolv.conf --settings=override --capability=CAP_NET_ADMIN --capability=CAP_NET_BIND_SERVICE --capability=CAP_SYS_ADMIN --capability=CAP_MAC_ADMIN --capability=CAP_SYS_MODULE
ExecStopPost={{.Path}}/dappctrl/post-stop.sh
KillMode=mixed
Type=notify
RestartForceExitStatus=133
SuccessExitStatus=133
# Enforce a strict device policy, similar to the one nspawn configures when it
# allocates its own scope unit. Make sure to keep these policies in sync if you
# change them!
DevicePolicy=closed
DeviceAllow=/dev/net/tun rwm
DeviceAllow=char-pts rw

# nspawn itself needs access to /dev/loop-control and /dev/loop, to implement
# the --image= option. Add these here, too.
DeviceAllow=/dev/loop-control rw
DeviceAllow=block-loop rw
DeviceAllow=block-blkext rw

# nspawn can set up LUKS encrypted loopback files, in which case it needs
# access to /dev/mapper/control and the block devices /dev/mapper/*.
DeviceAllow=/dev/mapper/control rw
DeviceAllow=block-device-mapper rw


[Install]
WantedBy=machines.target
