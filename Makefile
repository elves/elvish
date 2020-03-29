PKG_BASE := github.com/elves/elvish
PKGS := $(shell go list ./... | sed 's|^$(PKG_BASE)|.|')
PKG_COVERS := $(shell go list ./... | sed 's|^$(PKG_BASE)|.|' | grep -v '^\.$$' | sed 's/^\./cover/' | sed 's/$$/.cover/')
COVER_MODE := set
VERSION := $(shell git describe --tags --always --dirty=-dirty)

GOVERALLS := github.com/mattn/goveralls

default: test get

get:
	go get -trimpath -ldflags \
		"-X github.com/elves/elvish/pkg/buildinfo.Version=$(VERSION) \
		 -X github.com/elves/elvish/pkg/buildinfo.Reproducible=true" .

buildall:
	./tools/buildall.sh

generate:
	go generate ./...

test:
	go test -race $(PKGS)

cover/%.cover: %
	mkdir -p $(dir $@)
	go test -coverprofile=$@ -covermode=$(COVER_MODE) ./$<

cover/all: $(PKG_COVERS)
	echo mode: $(COVER_MODE) > $@
	for f in $(PKG_COVERS); do test -f $$f && sed 1d $$f >> $@ || true; done

upload-coverage-codecov: cover/all
	curl -s https://codecov.io/bash -o codecov.bash && \
		bash codecov.bash -f $< || \
		true

upload-coverage-coveralls: cover/all
	go get $(GOVERALLS)
	goveralls -coverprofile $<

binaries-travis:
	./tools/binaries-travis.sh

.PHONY: default get buildall generate test testmain upload-coverage-codecov upload-coverage-coveralls binaries-travis
