#!/bin/sh -e

# buildall.sh $SRC_DIR $DST_DIR $SUFFIX
#
# Builds Elvish binaries for all supported platforms, using the code in $SRC_DIR
# and building $DST_DIR/$GOOS-$GOARCH/elvish-$SUFFIX for each supported
# combination of $GOOS and $GOARCH.
#
# It also creates and an archive for each binary file, and puts it in the same
# directory. For GOOS=windows, the archive is a .zip file. For all other GOOS,
# the archive is a .tar.gz file.
#
# If the sha256sum command is available, this script also creates a sha256sum
# file for each binary and archive file, and puts it in the same directory.
#
# The ELVISH_REPRODUCIBLE environment variable, if set, instructs the script to
# mark the binary as a reproducible build. It must take one of the two following
# values:
#
# - release: SRC_DIR must contain the source code for a tagged release.
#
# - dev: SRC_DIR must be a Git repository checked out from the latest master
#        branch.
#
# This script is not whitespace-correct; avoid whitespaces in directory names.

if test $# != 3; then
    # Output the comment block above, stripping any leading "#" or "# "
    sed < $0 -n '
      /^# /,/^$/{
        /^$/q
        s/^# \?//
        p
      }'
    exit 1
fi

SRC_DIR=$1
DST_DIR=$2
SUFFIX=$3

LD_FLAGS=
if test -n "$ELVISH_REPRODUCIBLE"; then
    LD_FLAGS="-X src.elv.sh/pkg/buildinfo.Reproducible=true"
    if test "$ELVISH_REPRODUCIBLE" = dev; then
        LD_FLAGS="$LD_FLAGS -X src.elv.sh/pkg/buildinfo.VersionSuffix=-dev.$(git -C $SRC_DIR rev-parse HEAD)"
    elif test "$ELVISH_REPRODUCIBLE" = release; then
        : # nothing to do
    else
        echo "$ELVISH_REPRODUCIBLE must be 'dev' or 'release' when set"
    fi
fi

export GOOS GOARCH GOFLAGS
export CGO_ENABLED=0

main() {
    buildarch amd64 linux darwin freebsd openbsd netbsd windows
    buildarch 386   linux windows
    buildarch arm64 linux
}

# buildarch $arch $os...
#
# Builds one GOARCH, multiple GOOS.
buildarch() {
    local GOARCH=$1 GOOS
    shift
    for GOOS in $@; do
        buildone
    done
}

# buildone
#
# Builds one GOARCH and one GOOS.
#
# Uses: $GOARCH $GOOS $DST_DIR
buildone() {
    local BIN_DIR=$DST_DIR/$GOOS-$GOARCH
    mkdir -p $BIN_DIR

    local STEM=elvish-$SUFFIX
    if test $GOOS = windows; then
        local BIN=$STEM.exe
        local ARCHIVE=$STEM.zip
    else
        local BIN=$STEM
        local ARCHIVE=$STEM.tar.gz
    fi

    if test $GOOS = windows -o $GOOS = linux; then
        local GOFLAGS=-buildmode=pie
    fi

    echo -n "Building for $GOOS-$GOARCH... "
    go build -trimpath -ldflags "$LD_FLAGS"\
      -o $BIN_DIR/$BIN $SRC_DIR/cmd/elvish || {
        echo "Failed"
        return
    }

    (
    cd $BIN_DIR
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
