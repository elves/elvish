#!/bin/sh -e

# Should be invoked from repo root.
#
# Environment variables used by this script:
# $CIRRUS_BRANCH $CIRRUS_TAG $BINTRAY_TOKEN

# Used by buildall.sh
export ELVISH_REPRODUCIBLE

if [ "$CIRRUS_TAG" -a "$CIRRUS_BRANCH" != master ]; then
    file_suffix=$CIRRUS_TAG
    ELVISH_REPRODUCIBLE=release
else
	if [ "$CIRRUS_BRANCH" = master ]; then
        file_suffix=HEAD
    else
        file_suffix=$CIRRUS_BRANCH
	fi
    ELVISH_REPRODUCIBLE=dev
fi

rm -rf _bin
./tools/buildall.sh . _bin $file_suffix

bintray_version=$file_suffix
echo "Deleting old version of $bintray_version"
curl -X DELETE -u$BINTRAY_TOKEN \
    https://api.bintray.com/packages/elves/elvish/elvish/versions/$bintray_version

find _bin -type f | while read f; do
    f=${f#_bin/}
    echo "Deploying $f"
    curl -T _bin/$f -u$BINTRAY_TOKEN \
        https://api.bintray.com/content/elves/elvish/elvish/$bintray_version/$f'?publish=1&override=1'
done
