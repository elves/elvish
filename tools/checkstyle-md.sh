#!/bin/sh -e

# Check if the style of the Markdown files is correct without modifying those
# files.

# The `prettier` utility doesn't provide a "diff" option. Therefore, we have
# to do this in a convoluted fashion to get a diff from the current state to
# what `prettier` considers correct and reporting, via the exit status, that
# the check failed.

# We predicate the detailed diff on being in a CI environment since we don't
# care if the files are modified. If not, just list the pathnames that need to
# be reformatted without actually modifying those files.

if test "$CI" = ""
then
    echo 'Markdown files that need changes:'
    ! find . -name '*.md' |
        xargs prettier --tab-width 4 --prose-wrap always --list-different |
        sed 's/^/  /' | grep . && echo '  None!'
else
    echo 'Markdown files need these changes:'
    if ! find . -name '*.md' |
        xargs prettier --tab-width 4 --prose-wrap always --check >/dev/null
    then
        find . -name '*.md' |
            xargs prettier --tab-width 4 --prose-wrap always --write >/dev/null
        find . -name '*.md' | xargs git diff
        exit 1
    fi
    echo '  None!'
fi
