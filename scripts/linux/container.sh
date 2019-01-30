#!/bin/bash

container="privatix-dapp-container"
echo "Container name:" ${container}

dappcoredir="./app"
echo "App folder path:" ${dappcoredir}

# create container
sudo debootstrap jessie ./${container} http://mirror.yandex.ru/debian
#sudo chroot ${container} dpkg --print-architecture

# copy dappcore
sudo cp -R ${dappcoredir}/. ${container}/app/

# connect to container
sudo systemd-nspawn -D ${container}/ << EOF

#set root password: xHd26ksN
echo -e "xHd26ksN\nxHd26ksN\n" | passwd

echo "pts/0" >> /etc/securetty 
echo deb http://http.debian.net/debian jessie-backports main > /etc/apt/sources.list.d/jessie-backports.list

apt-get update
apt-get -t jessie-backports install -y systemd
apt-get install -y dbus

# install locale
apt-get install -y locales

echo "en_US.UTF-8 UTF-8" >> /etc/locale.gen
locale-gen
update-locale LANG=en_US.UTF-8

# install postgres
apt-get update
apt-get install -y ca-certificates

echo "deb [arch=amd64] http://apt.postgresql.org/pub/repos/apt/ jessie-pgdg main" > /etc/apt/sources.list.d/pgdg.list
wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
apt-get update
apt-get install -y postgresql-10

# configure postgres
service postgresql stop
sed -i.bup 's/port = 5432/port = 5433/g' /etc/postgresql/10/main/postgresql.conf
sed -i.bup 's/local.*all.*postgres.*peer/local all postgres trust/g' /etc/postgresql/10/main/pg_hba.conf
sed -i.buh 's/host.*all.*all.*127.0.0.1\/32.*md5/host all postgres 127.0.0.1\/32 trust/g' /etc/postgresql/10/main/pg_hba.conf
service postgresql start

# install Tor
apt-get install -y tor

# enable dappctrl daemon
mv /app/dappctrl/dappctrl.service /lib/systemd/system/
systemctl enable dappctrl.service

rm -rf /usr/share/doc/*
find /usr/share/locale -maxdepth 1 -mindepth 1 ! -name en_US -exec rm -rf {} \;
find /usr/share/i18n/locales -maxdepth 1 -mindepth 1 ! -name en_US -exec rm -rf {} \;
rm -rf /usr/share/man/*
rm -rf /usr/share/groff/*
rm -rf /usr/share/info/*
rm -rf /usr/share/lintian/*
rm -rf /usr/include/*

logout
EOF
    
# create tar archive
sudo tar cpJf ${container}.tar.xz --exclude="./var/cache/apt/archives/*.deb" \
--exclude="./var/lib/apt/lists/*" --exclude="./var/cache/apt/*.bin" \
--one-file-system -C ${container} .
