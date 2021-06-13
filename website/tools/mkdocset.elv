# Generate docset from reference docs.
#
# Docset is a format for packaging docs for offline consumption:
# https://kapeli.com/docsets
#
# External dependencies:
# 
# - python3
# - sqlite3

use path

if (!= 2 (count $args)) {
  echo "Usage: mkdocset.elv $website $docset"
  exit 1
}

var bindir = (path:dir (src)[name])
var website docset = $@args
var website-docs = $website/ref

mkdir -p $docset/Contents/Resources/Documents
cp $bindir/docset-data/Info.plist $docset/Contents
cp $website/ref/*.html $docset/Contents/Resources/Documents
rm $docset/Contents/Resources/Documents/{index language}.html
python3 $bindir/mkdsidx.py $website/ref | sqlite3 $docset/Contents/Resources/docSet.dsidx
