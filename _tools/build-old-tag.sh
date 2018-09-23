#!/bin/sh -e

TAG=$1
: ${BIN_DIR:=/data/bin}

REPO=github.com/elves/elvish
REPO_ADDR=https://$REPO

export GOPATH=`mktemp -d`
cleanup() {
    rm -rf $GOPATH
}
trap cleanup EXIT

git clone --depth=1 --branch $TAG $REPO_ADDR $GOPATH/src/$REPO
cp `dirname $0`/buildall.sh $GOPATH/src/$REPO
cd $GOPATH/src/$REPO
BIN_DIR=$BIN_DIR VERSION=$TAG ./buildall.sh
