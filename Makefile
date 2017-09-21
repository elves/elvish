PKG_BASE := github.com/elves/elvish
PKGS := $(shell go list ./... | grep -v /vendor/)
PKG_COVERS := $(shell go list ./... | grep -v '^$(PKG_BASE)/vendor/' | grep -v '^$(PKG_BASE)$$' | sed "s|^$(PKG_BASE)/|cover/|" | sed 's/$$/.cover/')
COVER_MODE := set
VERSION := $(shell git describe --tags --always)

FIRST_GOPATH=$(shell go env GOPATH | cut -d: -f1)

default: get test

get:
	go get -ldflags "-X main.Version=$(VERSION)" .

generate:
	go generate ./...

test:
	go test $(PKGS)

cover/%.cover: %
	mkdir -p $(dir $@)
	go test -coverprofile=$@ -covermode=$(COVER_MODE) ./$<

cover/all: $(PKG_COVERS)
	echo mode: $(COVER_MODE) > $@
	for f in $(PKG_COVERS); do test -f $$f && sed 1d $$f >> $@ || true; done

# We would love to test for coverage in pull requests, but it's now
# bettered turned off for two reasons:
#
# 1) The goverall badge will always show the "latest" coverage, even if that
# comes from a PR.
#
# 2) Some of the tests have fluctuating coverage (the test against
# edit.tty.AsyncReader), and goveralls will put a big cross on the PR when the
# coverage happens to drop.
goveralls: cover/all
	test "$(TRAVIS_PULL_REQUEST)" = false \
		&& go get -u github.com/mattn/goveralls \
		&& $(FIRST_GOPATH)/bin/goveralls -coverprofile=cover/all -service=travis-ci \
		|| echo "not sending to coveralls"

upload: get
	tar cfz elvish.tar.gz -C $(FIRST_GOPATH)/bin elvish
	echo "$(TRAVIS_GO_VERSION)" | grep -q '^1.8' \
		&& test "$(TRAVIS_PULL_REQUEST)" = false \
		&& test -n "$(TRAVIS_TAG)" -o "$(TRAVIS_BRANCH)" = master \
		&& curl http://ul.elvish.io:6060/ -F name=elvish-$(if $(TRAVIS_TAG),$(TRAVIS_TAG)-,)$(TRAVIS_OS_NAME).tar.gz \
			-F token=$$UPLOAD_TOKEN -F file=@./elvish.tar.gz\
		|| echo "not uploading"

travis: goveralls upload

.PHONY: default get generate test goveralls upload travis
