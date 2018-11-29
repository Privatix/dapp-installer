#!/usr/bin/env bash
DAPPINST=github.com/privatix/dapp-installer

echo ${DAPPINST_DIR:=${GOPATH}/src/${DAPPINST}}

if [ ! -f "${GOPATH}"/bin/dep ]; then
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
fi
echo running dep ensure
cd "${DAPPINST_DIR}" && dep ensure
go get -d ${DAPPINST_DIR}/...
go get -u github.com/rakyll/statik
go get -u github.com/josephspurrier/goversioninfo/cmd/goversioninfo
go get -u gopkg.in/reform.v1/reform

go generate ${DAPPINST_DIR}/...

GIT_COMMIT=$(git rev-list -1 HEAD)
GIT_RELEASE=$(git tag -l --points-at HEAD)

export GIT_COMMIT
export GIT_RELEASE

go build -ldflags "-X main.Commit=$GIT_COMMIT -X main.Version=$GIT_RELEASE" -tags=notest
