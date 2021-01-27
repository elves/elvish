#!/bin/sh -e

# Builds Elvish binaries for all supported platforms.
#
# Arguments are passed via the following environment variables:
#
# - SRC_DIR: Root of the checked out source tree. Defauls to ".".
#
# - BIN_DIR: Root of the directory to put binaries in. Defauls to "./_bin".
#
# - VERSION: Version info to use in the built binaries. Defaults to "unknown".
#
# - STEM_SUFFIX: Suffix to add to the filename stem. Defaults to $VERSION.
#
# Alternatively, they can be passed as command-line arguments (buildall.sh
# src_dir dst_dir version), but the environment variables take precedence.
#
# For each supported combination of $GOOS and $GOARCH, this script saves the
# built binary to $BIN_DIR/$GOOS-$GOARCH/elvish-$VERSION.
#
# It also creates and an archive for each binary file, and puts it in the same
# directory. For GOOS=windows, the archive is a .zip file. For all other GOOS,
# the archive is a .tar.gz file.
#
# If the sha256sum command is available, this script also creates a sha256sum
# file for each binary and archive file, and puts it in the same directory.
#
# This script is not whitespace-correct; avoid whitespaces in directory names.

: ${SRC_DIR:=${1:-.}}
: ${BIN_DIR:=${2:-./_bin}}
: ${VERSION:=${3:-unknown}}
: ${STEM_SUFFIX:=${4:-$VERSION}}

export GOOS GOARCH
export CGO_ENABLED=0

main() {
    buildarch amd64 linux darwin freebsd openbsd netbsd windows
    buildarch 386   linux windows
    buildarch arm64 linux
}

# Builds one GOARCH, multiple GOOS.
#
# buildarch $arch $os...
buildarch() {
    local GOARCH=$1 GOOS
    shift
    for GOOS in $@; do
        buildone
    done
}

# Builds one GOARCH and one GOOS.
#
# Uses: $GOARCH $GOOS $BIN_DIR
buildone() {
    local DST_DIR=$BIN_DIR/$GOOS-$GOARCH
    mkdir -p $DST_DIR

    local STEM=elvish-$STEM_SUFFIX
    if test $GOOS = windows; then
        local BIN=$STEM.exe
        local ARCHIVE=$STEM.zip
    else
        local BIN=$STEM
        local ARCHIVE=$STEM.tar.gz
    fi

    echo -n "Building for $GOOS-$GOARCH... "
    go build -o $DST_DIR/$BIN -trimpath -ldflags \
        "-X src.elv.sh/pkg/buildinfo.Version=$VERSION \
         -X src.elv.sh/pkg/buildinfo.Reproducible=true" $SRC_DIR/cmd/elvish || {
        echo "Failed"
        return
    }

    (
    cd $DST_DIR
    if test $GOOS = windows; then
        zip -q $ARCHIVE $BIN
    else
        tar cfz $ARCHIVE $BIN
    fi

    echo "Done"

    if which sha256sum > /dev/null; then
        sha256sum $BIN > $BIN.sha256sum
        sha256sum $ARCHIVE > $ARCHIVE.sha256sum
    fi
    )
}

main
