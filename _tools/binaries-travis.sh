#!/bin/sh -e

# Should be invoked from repo root. Required environment variables:
# $TRAVIS_BRANCH $BINTRAY_CREDENTIAL

if [ "$TRAVIS_BRANCH" = master ]; then
    export VERSION=HEAD
else
    export VERSION=$TRAVIS_BRANCH
fi
MANIFEST=_bin/manifest ./_tools/buildall.sh

cat _bin/manifest | while read f; do
    echo Deploying $f
    curl -T $f -u$BINTRAY_CREDENTIAL \
        https://api.bintray.com/content/elves/elvish/elvish/$VERSION/$f'?publish=1&override=1'
done
