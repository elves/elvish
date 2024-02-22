#!/usr/bin/env elvish

# This script is supposed to be run with Elvish at the same commit. Either
# ensure that Elvish is built and installed from the repo, or use "go run
# ./cmd/elvish tools/buildall.elv ...".

use flag
use os
use platform
use str

var platforms = [
  [&arch=amd64 &os=linux]
  [&arch=amd64 &os=darwin]
  [&arch=amd64 &os=freebsd]
  [&arch=amd64 &os=openbsd]
  [&arch=amd64 &os=netbsd]
  [&arch=amd64 &os=windows]

  [&arch=386 &os=linux]
  [&arch=386 &os=windows]

  [&arch=arm64 &os=linux]
  [&arch=arm64 &os=darwin]
]

var usage = ^
'buildall.elv [-name name] [-variant variant] go-pkg dst-dir

Builds $go-pkg, outputting $dst-dir/$GOOS-$GOARCH/$name and an archive for a
predefined list of supported GOOS/GOARCH combinations.

For GOOS=windows, the binary name has an .exe suffix, and the archive is a
.zip file. For all other GOOS, the archive is a .tar.gz file.

If the sha256sum command is available, this script also creates a sha256sum
file for each binary and archive file, and puts it in the same directory.

The value of $variant will be used to override
src.elv.sh/pkg/buildinfo.BuildVariant.
'

fn main {|go-pkg dst-dir &name=elvish &variant=''|
  tmp E:CGO_ENABLED = 0
  for platform $platforms {
    var arch os = $platform[arch] $platform[os]

    var bin-dir = $dst-dir/$os'-'$arch
    os:mkdir-all $bin-dir

    var bin-name archive-name = (
      if (eq $os windows) {
        put $name{.exe .zip}
      } else {
        put $name{'' .tar.gz}
      })

    print 'Building for '$os'-'$arch'... '

    tmp E:GOOS E:GOARCH = $os $arch

    try {
      go build ^
        -trimpath ^
        -ldflags '-X src.elv.sh/pkg/buildinfo.BuildVariant='$variant ^
        -o $bin-dir/$bin-name ^
        $go-pkg
    } catch e {
      echo 'Failed'
      continue
    }

    # This is needed to get files appear in the root of the archive files.
    tmp pwd = $bin-dir
    # Archive files store the modification timestamp of files. Change it to a
    # fixed point in time to make the archive files reproducible.
    touch -d 2000-01-01T00:00:00Z $bin-name
    if (eq $os windows) {
      zip -q $archive-name $bin-name
    } else {
      # If we create a .tar.gz file directly with the tar command, the
      # resulting archive will contain the timestamp of the .tar file,
      # rendering the result unreproducible. As a result, we need to do it in
      # two steps.
      tar cf $bin-name.tar $bin-name
      touch -d 2022-01-01T00:00:00Z $bin-name.tar
      gzip -f $bin-name.tar
    }
    # Update the modification time again to reflect the actual modification
    # time. (Technically this makes the file appear slightly newer han it really
    # is, but it's close enough).
    touch $bin-name
    echo 'Done'

    if (has-external sha256sum) {
      for file [$bin-name $archive-name] {
        sha256sum $file > $file.sha256sum
      }
    }
  }
}

flag:call $main~ $args &on-parse-error={|_| print $usage; exit 1}
