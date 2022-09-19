#!/bin/sh

# Check if the source code has any spelling errors.
status=0

x="$(codespell)"
if [ "$x" != "" ]; then
    echo
    echo '========================================'
    echo 'Spelling errors identified by codepsell:'
    echo '========================================'
    echo
    echo "$x"
    echo
    status=1
fi

if test "$CI" != ""; then
    exit $status
fi
exit 0
