#!/bin/sh -e

# Verify the style of the Go source files without modifying them.
status=0

x="$(find . -name '*.go' | xargs goimports -d)"
if [ "$x" != "" ]; then
    echo
    echo '==================================='
    echo 'Go files need these import changes:'
    echo '==================================='
    echo
    echo "$x"
    echo
    status=1
fi

x="$(find . -name '*.go' | xargs gofmt -s -d)"
if [ "$x" != "" ]; then
    echo
    echo '==========================================='
    echo 'Go files need that need formatting changes:'
    echo '==========================================='
    echo
    echo "$x"
    echo
    status=1
fi

if test "$CI" != ""; then
    exit $status
fi
exit 0
