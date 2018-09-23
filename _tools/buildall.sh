#!/bin/sh -e

# Should be invoked from repo root.

: ${VERSION:=unknown}
: ${BIN_DIR:=./_bin}
: ${MANIFEST:=/dev/null}

export GOOS GOARCH
export CGO_ENABLED=0

# build $os $arch...
build() {
    local GOARCH=$1
    shift
    for GOOS in $@; do
        DST_DIR=$BIN_DIR/$GOOS-$GOARCH
        mkdir -p $DST_DIR
        buildone
    done
}

# buildone
# Uses: $GOOS $GOARCH $DST_DIR
buildone() {
    STEM=elvish-$VERSION
    if test $GOOS = windows; then
        BIN=$STEM.exe
        ARCHIVE=$STEM.zip
    else
        BIN=$STEM
        ARCHIVE=$STEM.tar.gz
    fi

    echo "Building for $GOOS-$GOARCH"
    go build -o $DST_DIR/$BIN -ldflags \
        "-X github.com/elves/elvish/buildinfo.Version=$VERSION \
         -X github.com/elves/elvish/buildinfo.GoRoot=`go env GOROOT` \
         -X github.com/elves/elvish/buildinfo.GoPath=`go env GOPATH`" || {
        echo "  -> Failed"
        return
    }

    (
    cd $DST_DIR
    if test $GOOS = windows; then
        zip -q $ARCHIVE $BIN
    else
        tar cfz $ARCHIVE $BIN
    fi
    )

    echo " -> Done"
    echo $DST_DIR/$ARCHIVE >> $MANIFEST
}

build amd64 linux darwin freebsd openbsd netbsd windows
build 386   linux windows
build arm64 linux
