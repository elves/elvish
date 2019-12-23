#!/bin/sh -e

# Should be invoked from repo root. Required environment variables:
# $TRAVIS_BRANCH $BINTRAY_CREDENTIAL

# Manipulate the GOROOT so that it does not contain the Go version number
ln -s `go env GOROOT` $TRAVIS_HOME/goroot
export GOROOT=$TRAVIS_HOME/goroot

if [ "$TRAVIS_BRANCH" = master ]; then
    export VERSION=HEAD
else
    export VERSION=$TRAVIS_BRANCH
fi
MANIFEST=_bin/manifest ./tools/buildall.sh

echo "Deleting old HEAD"
curl -X DELETE -u$BINTRAY_CREDENTIAL \
    https://api.bintray.com/packages/elves/elvish/elvish/versions/$VERSION

cat _bin/manifest | while read f; do
    echo "Deploying $f"
    curl -T _bin/$f -u$BINTRAY_CREDENTIAL \
        https://api.bintray.com/content/elves/elvish/elvish/$VERSION/$f'?publish=1&override=1'
done
