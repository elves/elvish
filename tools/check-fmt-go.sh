#!/bin/sh -e
# Check if Go files are properly formatted without modifying them.

echo 'Go files need these changes:'
# The grep is needed because `goimports -d` and `gofmt -d` always exits with 0.
if find . -name '*.go' | xargs goimports -d | grep .; then
    exit 1
fi
if find . -name '*.go' | xargs gofmt -s -d | grep .; then
    exit 1
fi
echo '  None!'
