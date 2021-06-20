#!/bin/sh -e

# Prune the same objects from the "make cover" report that we tell Codecov
# (https://codecov.io/gh/elves/elvish/) to ignore.

if test $# != 2
then
    echo 'Usage: cover_prune.sh ${codecov.yml} $cover' >&2
    exit 1
fi
yaml="$1"
data="$2"

sed -En '/^ignore:/,/^[^ ]/s/^  *- "(.*)"/src.elv.sh\/\1/p' $yaml > $yaml.ignore
grep -F -v -f $yaml.ignore $data > $data.pruned
mv $data.pruned $data
rm $yaml.ignore
