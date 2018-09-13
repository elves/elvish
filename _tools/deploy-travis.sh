#!/bin/sh -e

# Should be invoked from repo root. Required environment variables:
# $TRAVIS_BRANCH $BINTRAY_CREDENTIAL

mkdir -p bin
if [ "$TRAVIS_BRANCH" = master ]; then
    export VERSION=HEAD
else
    export VERSION=$TRAVIS_BRANCH
fi
./_tools/buildall.sh

cd bin
cat manifest | while read f; do
    echo Deploying $f
    curl -T $f -uxiaq:$BINTRAY_CREDENTIAL \
        https://api.bintray.com/content/elves/elvish/elvish/$VERSION/$f'?publish=1'
done
