#!/bin/sh
# Check Go source files for disallowed content.

ret=0

# We have our own trimmed-down copy of net/rpc to reduce binary size. Make sure
# that dependency on net/rpc is not accidentally introduced.
x=$(find . -name '*.go' | xargs grep '"net/rpc"')
if test "$x" != ""; then
    echo "======================================================================"
    echo 'Disallowed import of net/rpc:'
    echo "======================================================================"
    echo "$x"
    ret=1
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
    ret=1
fi

# doc:show depends on references to language.html to not have a ./ prefix.
x=$(find . -name '*.elv' | xargs grep '(\.\/language\.html')
if test "$x" != ""; then
    echo "======================================================================"
    echo 'Disallowed use of ./ in link destination to language.html'
    echo "======================================================================"
    echo "$x"
    ret=1
fi

exit $ret
