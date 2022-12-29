#!/bin/sh
# Check that generated Go source files are up to date.

git_unstaged() {
    # The output of "git status -s" starts with two letters XY, where Y is the
    # status in the working tree. Files that are staged in the index have Y
    # being a space; exclude them.
    git status -s | grep '^.[^ ]'
}

if ! which git >/dev/null; then
    echo "$0 requires Git"
    exit 1
fi

if test "$(git_unstaged)" != ""; then
    echo "$0 must be run from a Git repo with no unstaged changes or untracked files"
    exit 1
fi

go generate ./... || exit 1
x=$(git_unstaged)

if test "$x" != ""; then
    echo "======================================================================"
    echo "Generated Go code is out of date. See"
    echo "https://github.com/elves/elvish/blob/master/CONTRIBUTING.md#generated-code"
    echo "======================================================================"
    echo "$x"
    exit 1
fi
