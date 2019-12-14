PKG_BASE := github.com/elves/elvish
PKGS := $(shell go list ./... | sed 's|^$(PKG_BASE)|.|')
PKG_COVERS := $(shell go list ./... | sed 's|^$(PKG_BASE)|.|' | grep -v '^\.$$' | sed 's/^\./_cover/' | sed 's/$$/.cover/')
COVER_MODE := set
VERSION := $(shell git describe --tags --always --dirty=-dirty)

GOVERALLS := github.com/mattn/goveralls

default: test get

# TODO(xiaq): Add -trimpath when we require Go >= 1.13.
get:
	go get -ldflags "-X github.com/elves/elvish/buildinfo.Version=$(VERSION)" .

buildall:
	./_tools/buildall.sh

generate:
	go generate ./...

test:
	go test $(PKGS)

_cover/%.cover: %
	mkdir -p $(dir $@)
	go test -coverprofile=$@ -covermode=$(COVER_MODE) ./$<

_cover/all: $(PKG_COVERS)
	echo mode: $(COVER_MODE) > $@
	for f in $(PKG_COVERS); do test -f $$f && sed 1d $$f >> $@ || true; done

upload-coverage-codecov: _cover/all
	curl -s https://codecov.io/bash -o codecov.bash
	bash codecov.bash -f $<

upload-coverage-coveralls: _cover/all
	go get $(GOVERALLS)
	goveralls -coverprofile $<

binaries-travis:
	./_tools/binaries-travis.sh

.PHONY: default get buildall generate test testmain upload-coverage-codecov upload-coverage-coveralls binaries-travis
