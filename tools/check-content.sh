#!/bin/sh -e

# Check Go source files for disallowed content.

echo 'Disallowed import of net/rpc:'
if find . -name '*.go' | xargs grep '"net/rpc"'; then
    exit 1
else
    echo '  None!'
fi

echo 'Disallowed call of unix.{Umask,Getrlimit,Setrlimit}'
if find . -name '*.go' | egrep -v '\./pkg/(mods/unix|daemon|testutil)' | xargs egrep 'unix\.(Umask|Getrlimit|Setrlimit)'; then
    exit 1
else
    echo '  None!'
fi
