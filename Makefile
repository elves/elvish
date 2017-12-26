PKG_BASE := github.com/elves/elvish
PKGS := $(shell go list ./... | sed 's|^$(PKG_BASE)|.|' | grep -v '^./vendor')
PKG_COVERS := $(shell go list ./... | sed 's|^$(PKG_BASE)|.|' | grep -v '^\./vendor' | grep -v '^\.$$' | sed 's/^\./cover/' | sed 's/$$/.cover/')
COVER_MODE := set
VERSION := $(shell git describe --tags --always)

GOVERALLS := github.com/mattn/goveralls

default: test get

get:
	go get -ldflags "-X github.com/elves/elvish/build.Version=$(VERSION) -X github.com/elves/elvish/build.Builder=$(shell id -un)@$(shell hostname)" .

generate:
	go generate ./...

test:
	go test $(PKGS)

testmain:
	go test .

cover/%.cover: %
	mkdir -p $(dir $@)
	go test -coverprofile=$@ -covermode=$(COVER_MODE) ./$<

cover/all: $(PKG_COVERS)
	echo mode: $(COVER_MODE) > $@
	for f in $(PKG_COVERS); do test -f $$f && sed 1d $$f >> $@ || true; done

# Disable coverage reports for pull requests. The general testability of the
# code is pretty bad and it is premature to require contributors to maintain
# code coverage.
upload-codecov-travis: cover/all
	test "$(TRAVIS_PULL_REQUEST)" = false \
		&& echo "$(TRAVIS_GO_VERSION)" | grep -q '^1.9' \
		&& curl -s https://codecov.io/bash -o codecov.bash \
		&& bash codecov.bash -f $< \
		|| echo "not sending to codecov.io"

upload-coveralls-travis: cover/all
	test "$(TRAVIS_PULL_REQUEST)" = false \
		&& echo "$(TRAVIS_GO_VERSION)" | grep -q '^1.9' \
		&& go get -d $(GOVERALLS) \
		&& go build -o goveralls $(GOVERALLS) \
		&& ./goveralls -coverprofile $< -service=travis-ci \
		|| echo "not sending to coveralls"


upload-codecov-appveyor: cover/all
	codecov -f $<

upload-coveralls-appveyor: cover/all
	goveralls -coverprofile $< -service=appveyor-ci \
		|| echo "failed to upload to coveralls"

upload-bin:
	test "$(TRAVIS_OS_NAME)" = linux \
		&& echo "$(TRAVIS_GO_VERSION)" | grep -q '^1.9' \
		&& test "$(TRAVIS_PULL_REQUEST)" = false \
		&& test -n "$(TRAVIS_TAG)" -o "$(TRAVIS_BRANCH)" = master \
		&& go build -o ./elvish \
		&& ./elvish build-and-upload.elv \
		|| echo "not build-and-uploading"

travis: testmain upload-codecov-travis upload-coveralls-travis upload-bin
appveyor: testmain upload-codecov-appveyor upload-coveralls-appveyor

.PHONY: default get generate test testmain upload-codecov-travis upload-coveralls-travis upload-codecov-appveyor upload-coveralls-appveyor upload-bin travis
