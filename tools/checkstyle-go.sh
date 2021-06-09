#!/bin/sh -e

# Check if the style of the Go source files is correct without modifying those
# files.

# The grep is needed because `goimports -d` and `gofmt -d` always exits with 0.

echo 'Go files need these changes:'
if find . -name '*.go' | xargs goimports -d | grep .; then
    exit 1
fi
if find . -name '*.go' | xargs gofmt -s -d | grep .; then
    exit 1
fi
echo '  None!'
