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

if test "$CI" = ""; then
    # Presumably we're being run interactively by a developer.
    x="$(find . -name '*.md' | xargs prettier --list-different || true)"
    if test "$x" != ""; then
        echo
        echo '==========================================='
        echo 'Markdown files that need formatting changes'
        echo '(run "make style" to fix the problems):'
        echo '==========================================='
        echo
        echo "$x"
        echo
    fi
    exit 0
fi

if ! find . -name '*.md' | xargs prettier --check >/dev/null; then
    echo
    echo '=================================='
    echo 'Markdown files need these changes:'
    echo '=================================='
    echo
    find . -name '*.md' | xargs prettier --write >/dev/null
    find . -name '*.md' | xargs git diff
    echo
    exit 1
fi
