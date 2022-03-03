#!/bin/sh
# Prints "-race" if running on a platform that supports the race detector.
# This should be kept in sync with the official list here:
# https://golang.org/doc/articles/race_detector#Requirements
if test `go env CGO_ENABLED` = 1; then
  if echo `go env GOOS GOARCH` |
     egrep -qx '((linux|darwin|freebsd|netbsd) amd64|(linux|darwin) arm64|linux ppc64le)'; then
    printf %s -race
  elif echo `go env GOOS GOARCH` | egrep -qx 'windows amd64'; then
    # Race detector on Windows AMD64 requires gcc: https://github.com/golang/go/issues/27089
    if which gcc > /dev/null 2>&1; then
      printf %s -race
    fi
  fi
fi
