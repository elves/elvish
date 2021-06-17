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

# This approach to ignoring code from the "make cover" report is suboptimal
# but efficient enough for our purposes. Especially when weighed against a
# more complicated approach that constructs a single regexp that is used to
# filter the data just once since the number of exclusions is small.
sed -ne '/^ignore:/,/^[^ \t]/s/^[ \t]*- "\(.*\)"/\1/p' $yaml |
    while read pattern
    do
        grep -v "$pattern" $data > $data.tmp
        mv $data.tmp $data
    done
