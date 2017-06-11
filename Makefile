PKG_BASE := github.com/xiaq/persistent
PKGS := $(shell go list ./... | grep -v /vendor/)
PKG_COVERS := $(shell go list ./... | grep -v '^$(PKG_BASE)/vendor/' | grep -v '^$(PKG_BASE)$$' | sed "s|^$(PKG_BASE)/|cover/|" | sed 's/$$/.cover/')
COVER_MODE := set

FIRST_GOPATH=$(shell go env GOPATH | cut -d: -f1)

default: get test

get:
	go get .

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

travis: goveralls

.PHONY: default get generate test goveralls upload travis