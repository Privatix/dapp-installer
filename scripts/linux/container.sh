#!/bin/bash

container="privatix-dapp-container"
echo "Container name:" ${container}

dappcoredir="./app"
echo "App folder path:" ${dappcoredir}

# create container
sudo debootstrap jessie ./${container} http://mirror.yandex.ru/debian
#sudo chroot ${container} dpkg --print-architecture

# connect to container
sudo systemd-nspawn -D ${container}/ << EOF

#set root password: xHd26ksN
echo -e "xHd26ksN\nxHd26ksN\n" | passwd

# install locale
apt-get update
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

# create folder structure
mkdir -p /app
mkdir -p /app/dappctrl
mkdir -p /app/dappgui
mkdir -p /app/log
mkdir -p /app/product

exit
EOF

# copy dappcore
sudo cp -R ${dappcoredir}/. ${container}/app/

# move daemons
sudo mv ${container}/app/daemon/* ${container}/lib/systemd/system/

# sudo systemd-nspawn -D ${container}/ << EOF
# # # enable daemon
# # systemctl enable dappctrl.service

# # # create, migrate and init database
# # /app/dappctrl/dappctrl db-create -conn "dbname=postgres host=127.0.0.1 user=postgres sslmode=disable port=5433"
# # /app/dappctrl/dappctrl db-migrate -conn "dbname=dappctrl host=127.0.0.1 user=postgres sslmode=disable port=5433"
# # /app/dappctrl/dappctrl db-init-data -conn "dbname=dappctrl host=127.0.0.1 user=postgres sslmode=disable port=5433"

# exit
# EOF

# # create tar archive
# sudo tar cpJf ${container}.tar.xz --one-file-system -C ${container} .
