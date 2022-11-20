#!/bin/sh -e

# Check if the style of the Markdown files is correct without modifying those
# files.

echo 'Markdown files that need changes:'
if find . -name '*.md' | grep -v '/node_modules/' | xargs go run src.elv.sh/cmd/elvmdfmt -width 80 -d | grep .; then
    exit 1
fi
echo '  None!'
