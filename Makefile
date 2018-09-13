PKG_BASE := github.com/elves/elvish
PKGS := $(shell go list ./... | sed 's|^$(PKG_BASE)|.|' | grep -v '^./\(vendor\|website\)')
PKG_COVERS := $(shell go list ./... | sed 's|^$(PKG_BASE)|.|' | grep -v '^\./\(vendor\|website\)' | grep -v '^\.$$' | sed 's/^\./_cover/' | sed 's/$$/.cover/')
COVER_MODE := set
VERSION := $(shell git describe --tags --always --dirty=-dirty)

GOVERALLS := github.com/mattn/goveralls

default: test get

get:
	go get -ldflags "-X github.com/elves/elvish/buildinfo.Version=$(VERSION) \
		-X github.com/elves/elvish/buildinfo.GoRoot=$(shell go env GOROOT) \
		-X github.com/elves/elvish/buildinfo.GoPath=$(shell go env GOPATH)" .

buildall:
	./_tools/buildall.sh

generate:
	go generate ./...

test:
	go test $(PKGS)

testmain:
	go test .

_cover/%.cover: %
	mkdir -p $(dir $@)
	go test -coverprofile=$@ -covermode=$(COVER_MODE) ./$<

_cover/all: $(PKG_COVERS)
	echo mode: $(COVER_MODE) > $@
	for f in $(PKG_COVERS); do test -f $$f && sed 1d $$f >> $@ || true; done

upload-codecov-travis: _cover/all
	curl -s https://codecov.io/bash -o codecov.bash \
		&& bash codecov.bash -f $< \
		|| echo "not sending to codecov.io"

upload-coveralls-travis: _cover/all
	go get -d $(GOVERALLS) \
		&& go build -o goveralls $(GOVERALLS) \
		&& ./goveralls -coverprofile $< -service=travis-ci \
		|| echo "not sending to coveralls.io"

# Disable coverage reports for pull requests. The general testability of the
# code is pretty bad and it is premature to require contributors to maintain
# code coverage.

upload-codecov-appveyor: _cover/all
	test -z "$(APPVEYOR_PULL_REQUEST_NUMBER)" \
		&& codecov -f $< \
		|| echo "not sending to codecov.io"

upload-coveralls-appveyor: _cover/all
	test -z "$(APPVEYOR_PULL_REQUEST_NUMBER)" \
		&& goveralls -coverprofile $< -service=appveyor-ci \
		|| echo "not sending to coveralls.io"

deploy-travis:
	./_tools/deploy-travis.sh

coverage-travis: upload-codecov-travis upload-coveralls-travis
appveyor: testmain upload-codecov-appveyor upload-coveralls-appveyor

.PHONY: default get buildall generate test testmain upload-codecov-travis upload-coveralls-travis upload-codecov-appveyor upload-coveralls-appveyor coverage-travis appveyor deploy-travis
