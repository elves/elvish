#!/bin/sh -e

# Check if the style of the Go source files is correct without modifying those
# files.

# The grep is because `goimports -d` will exit with a non-zero status even
# when it doesn't suggest any changes. The grep will return failure if there
# is no output from goimports and sucess if there are any proposed changes.

echo 'Go files need these changes:'
if find . -name '*.go' | xargs goimports -d | grep .
then
    exit 1
fi
echo '  None!'
