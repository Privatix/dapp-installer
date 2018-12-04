#!/usr/bin/env bash
DAPPINST=github.com/privatix/dapp-installer

echo ${DAPPINST_DIR:=${GOPATH}/src/${DAPPINST}}

if [ ! -f "${GOPATH}"/bin/dep ]; then
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
fi

echo
echo dep ensure
echo

cd "${DAPPINST_DIR}" && dep ensure -v

echo
echo go get
echo

go get -d -v ${DAPPINST}/...
go get -u -v github.com/rakyll/statik
go get -u -v github.com/josephspurrier/goversioninfo/cmd/goversioninfo
go get -u -v github.com/denisbrodbeck/machineid
go get -u -v gopkg.in/reform.v1/reform

echo
echo go generate
echo

go generate -x ${DAPPINST}/...

GIT_COMMIT=$(git rev-list -1 HEAD)
GIT_RELEASE=$(git tag -l --points-at HEAD)

echo
echo go build
echo

echo $GOPATH/bin/dapp-installer
go build -o $GOPATH/bin/dapp-installer \
-ldflags "-X main.Commit=$GIT_COMMIT -X main.Version=$GIT_RELEASE" -tags=notest || exit 1

echo
echo done
