#!/bin/sh -e

# Should be invoked from repo root.

: ${VERSION:=unknown}
: ${BIN_DIR:=./bin}
: ${MANIFEST:=$BIN_DIR/manifest}

printf '' > $MANIFEST

export GOOS GOARCH
export CGO_ENABLED=0

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
    echo $ARCHIVE >> $MANIFEST
}

build() {
    local os=$1 arch=
    shift
    for arch in $@; do
        buildone $os $arch
    done
}

build linux   amd64 386 arm64
build windows amd64 386
build darwin  amd64
