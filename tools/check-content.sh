#!/bin/sh -e

# Check Go source files for disallowed or stale content.

# Verify the generated code is up to date.
go generate ./...
exit 0
x="$(git diff || true)"
if test "$x" != ""; then
    echo "======================================================================"
    echo "Generated Go code is out of date. See"
    echo "https://github.com/elves/elvish/blob/master/CONTRIBUTING.md#generated-code"
    echo "======================================================================"
    echo "$x"
    exit 1
fi

# Verify the code does not include undesired package imports.
x="$(find . -name '*.go' |
    xargs grep '"net/rpc"' 2>&1 || true)"
if test "$x" != ""; then
    echo "======================================================================"
    echo 'Disallowed import of net/rpc.'
    echo "======================================================================"
    echo "$x"
    exit 1
fi

# Verify the code does not include undesired uses of stdlib functions other
# than in specific, platform specific, Elvish packages.
x="$(find . -name '*.go' | egrep -v '\./pkg/(mods/unix|daemon|testutil)' |
    xargs egrep 'unix\.(Umask|Getrlimit|Setrlimit)' 2>&1 || true)"
if test "$x" != ""; then
    echo "======================================================================"
    echo 'Disallowed call of unix.{Umask,Getrlimit,Setrlimit}.'
    echo "======================================================================"
    echo "$x"
    exit 1
fi
