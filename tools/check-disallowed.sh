#!/bin/sh
# Check Go source files for disallowed content.

# We have our own trimmed-down copy of net/rpc to reduce binary size. Make sure
# that dependency on net/rpc is not accidentally introduced.
x=$(find . -name '*.go' | xargs grep '"net/rpc"')
if test "$x" != ""; then
    echo "======================================================================"
    echo 'Disallowed import of net/rpc:'
    echo "======================================================================"
    echo "$x"
    exit 1
fi

# The correct functioning of the unix: module depends on some certain calls not
# being made elsewhere.
x=$(find . -name '*.go' | egrep -v '\./pkg/(mods/unix|daemon|testutil)' |
    xargs egrep 'unix\.(Umask|Getrlimit|Setrlimit)')
if test "$x" != ""; then
    echo "======================================================================"
    echo 'Disallowed call of unix.{Umask,Getrlimit,Setrlimit}:'
    echo "======================================================================"
    echo "$x"
    exit 1
fi
