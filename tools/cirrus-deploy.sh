#!/bin/sh -e

# Should be invoked from repo root.
#
# Environment variables used by this script:
# $CIRRUS_BRANCH $CIRRUS_TAG $BINTRAY_TOKEN

if [ "$CIRRUS_BRANCH" = master ]; then
    export VERSION=HEAD
elif [ "$CIRRUS_TAG " ]; then
    export VERSION=$CIRRUS_TAG
else
    export VERSION=$CIRRUS_BRANCH
fi
MANIFEST=_bin/manifest ./tools/buildall.sh

echo "Deleting old HEAD"
curl -X DELETE -u$BINTRAY_TOKEN \
    https://api.bintray.com/packages/elves/elvish/elvish/versions/$VERSION

cat _bin/manifest | while read f; do
    echo "Deploying $f"
    curl -T _bin/$f -u$BINTRAY_TOKEN \
        https://api.bintray.com/content/elves/elvish/elvish/$VERSION/$f'?publish=1&override=1'
done
