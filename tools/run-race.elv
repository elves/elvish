#!/usr/bin/env elvish
# Prints "-race" if running on a platform that supports the race detector.
use re

# Keep in sync with the official list here:
# https://golang.org/doc/articles/race_detector#Requirements
var supported-os-arch = [
  linux/amd64
  linux/ppc64le
  linux/arm64
  linux/s390x
  freebsd/amd64
  netbsd/amd64
  darwin/amd64
  darwin/arm64
]

if (eq 1 (go env CGO_ENABLED)) {
  var os arch = (go env GOOS GOARCH)
  var os-arch = $os/$arch
  if (has-value $supported-os-arch $os-arch) {
    echo -race
  } elif (eq windows/amd64 $os-arch) {
    # Race detector on windows/amd64 requires gcc:
    # https://github.com/golang/go/issues/27089
    if (has-external gcc) {
      echo -race
    }
  }
}
