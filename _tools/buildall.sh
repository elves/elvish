#!/bin/sh -e

IFS=

VERSION=`git describe --tags --always --dirty=-dirty`

: ${BUILDER:=`id -un`@`hostname`}
: ${VERSION_SUFFIX:=$VERSION}
: ${BIN_DIR:=./bin}

export GOOS GOARCH

buildone() {
    GOOS=$1
    GOARCH=$2
    STEM=elvish-$GOOS-$GOARCH-$VERSION_SUFFIX
    if test $GOOS = windows; then
        BIN=$STEM.exe
        ARCHIVE=$STEM.zip
    else
        BIN=$STEM
        ARCHIVE=$STEM.tar.gz
    fi

    echo "Going to build $BIN"
    go build -o $BIN_DIR/$BIN -ldflags \
        "-X github.com/elves/elvish/buildinfo.Version=$VERSION \
         -X github.com/elves/elvish/buildinfo.GoRoot=`go env GOROOT` \
         -X github.com/elves/elvish/buildinfo.GoPath=`go env GOPATH`"

    (
    cd $BIN_DIR
    if test $GOOS = windows; then
        zip $ARCHIVE $BIN
    else
        tar cfz $ARCHIVE $BIN
    fi
    )

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
