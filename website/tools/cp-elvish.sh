#!/bin/sh
#
# This is a helper script used by the ttyshot generation program to copy just
# the non-hidden files in the top-level Elvish source directory to the ttyshot
# hermetic home directory. This creates a predictable directory for
# demonstrating things like "navigation" mode.
#
home="$1"
mkdir "$home/elvish"
for f in *
do
    if [ -d "$f" ]
    then
        mkdir "$home/elvish/$f"
    else
        cp "$f" "$home/elvish/$f"
    fi
done
