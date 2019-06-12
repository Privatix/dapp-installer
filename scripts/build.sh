#!/usr/bin/env bash
MY_PATH="`dirname \"$0\"`" # relative bash file path
DAPPINST_DIR="`( cd \"$MY_PATH/..\" && pwd )`"  # absolutized and normalized dappctrl path

echo ${DAPPINST_DIR}
cd "${DAPPINST_DIR}"

echo
echo go get
echo

go get -u -v github.com/rakyll/statik
go get -u -v github.com/josephspurrier/goversioninfo/cmd/goversioninfo
go get -u -v github.com/denisbrodbeck/machineid

echo
echo go generate
echo

go generate -x ${DAPPINST_DIR}/...

GIT_COMMIT=$(git rev-list -1 HEAD | head -n 1)
GIT_RELEASE=$(git tag -l --points-at HEAD | head -n 1)

# if $GIT_RELEASE is zero:
GIT_RELEASE=${GIT_RELEASE:-$(git rev-parse --abbrev-ref HEAD | grep -o "[0-9]\{1,\}\.[0-9]\{1,\}\.[0-9]\{1,\}")}


echo
echo go build
echo

if [[ ! -d "${GOPATH}/bin/" ]]; then
    mkdir "${GOPATH}/bin/" || exit 1
fi

echo $GOPATH/bin/dapp-installer

go build -o $GOPATH/bin/dapp-installer \
-ldflags "-X main.Commit=$GIT_COMMIT -X main.Version=$GIT_RELEASE" -tags=notest || exit 1

echo
echo done
