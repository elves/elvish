#!/bin/sh -e

IFS=

: ${BUILDER:=`id -un`@`hostname`}
: ${VERSION:=`git describe --tags --always`}
: ${BIN_DIR:=./bin}

export GOOS GOARCH

buildone() {
    GOOS=$1
    GOARCH=$2
    STEM=elvish-$GOOS-$GOARCH-$VERSION
    if test $GOOS = windows; then
        BIN=$STEM.exe
        ARCHIVE=$STEM.zip
    else
        BIN=$STEM
        ARCHIVE=$STEM.tar.gz
    fi

    echo "Going to build $BIN"
    go build -o $BIN_DIR/$BIN -ldflags "-X github.com/elves/elvish/build.Version=$VERSION -X github.com/elves/elvish/build.Builder=$BUILDER"

    if test $GOOS = windows; then
        zip $BIN_DIR/$ARCHIVE $BIN_DIR/$BIN
    else
        tar cfz $BIN_DIR/$ARCHIVE $BIN_DIR/$BIN
    fi
    echo "Built $BIN and archived $ARCHIVE"
}

build() {
    local os=$1 arch=
    shift
    for arch in $@; do
        buildone $os $arch
    done
}

build darwin  amd64
build windows amd64 386
build linux   amd64 386 arm64
