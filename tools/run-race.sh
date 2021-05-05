#!/bin/sh
# Prints "-race" if tests should be run with the race detector.
#
# This is a subset of all platforms that actually support the race detector,
# which is documented in
# https://golang.org/doc/articles/race_detector#Supported_Systems.
#
# We don't run with race detectors on Windows because it requires GCC, which is
# not always available.
if echo `go env GOOS GOARCH CGO_ENABLED` |
   egrep -qx '((linux|darwin|freebsd|netbsd) amd64|(linux|darwin) arm64) 1'; then
  printf %s -race
fi
