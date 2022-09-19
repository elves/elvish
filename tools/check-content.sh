#!/bin/sh

# Check Go source files for disallowed content.
status=0

x="$(find . -name '*.go' | xargs grep '"net/rpc"')"
if [ "$x" != "" ]; then
    echo
    echo '==================================='
    echo 'Go files should not import net/rpc:'
    echo '==================================='
    echo
    echo "$x"
    status=1
fi

x="$(find . -name '*.go' |
        egrep -v '\./pkg/(mods/unix|daemon|testutil)' |
        xargs egrep 'unix\.(Umask|Getrlimit|Setrlimit)')"
if test "$x" != ""; then
    echo
    echo '==================================================='
    echo 'Disallowed call of unix.{Umask,Getrlimit,Setrlimit}'
    echo '==================================================='
    echo
    echo "$x"
    status=1
fi

if test "$CI" != ""; then
    exit $status
fi
exit 0
