#!/bin/bash -e
#
# This script exists due to the problem noted in
# https://github.com/elves/elvish/issues/1187. Specifically, `go test -cover`
# only instruments code in the same package as the unit test. Which
# effectively excludes coverage by integration tests.
#
# Testing each package individually with `-coverpackage=./...` produces, in
# aggregate, a more accurate picture of which statements are executed by all
# tests.

# Exclude these test utilities (that will always have low coverage), and other
# packages we don't care about having low coverage (e.g., pkg/web), from test
# coverage reports.
excluded_modules=(
  "pkg/cli/clitest"
  "pkg/eval/evaltest"
  "pkg/prog/progtest"
  "pkg/store/storetest"
  "pkg/web"
)

exclusion_file=$(mktemp)
rm -f $exclusion_file
for excluded in ${excluded_modules[*]}
do
    echo "$excluded" >>$exclusion_file
done

rm -f coverage
echo 'mode: set' >coverage
for pkg in $(go list ./...)
do
    exclude=false
    for excluded in ${excluded_modules[*]}
    do
        if [[ $pkg == *"$excluded"* ]]; then
            exclude=true
            break
        fi
    done
    if [[ $exclude == true ]]
    then
        continue
    fi

    # The `|| true` is needed because some packages may not have any test
    # coverage. We don't want that to cause this script to terminate.
    rm -f coverage.tmp
    go test -covermode=set -coverprofile=coverage.tmp -coverpkg=./... $pkg || true
    tail +2 coverage.tmp | grep -v -F -f $exclusion_file >>coverage || true
done
rm -f coverage.tmp $exclusion_file
