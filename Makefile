PKGS := $(shell go list ./... | grep -v /vendor/)
PKG_COVERS := $(shell go list ./... | grep -v /vendor/ | grep "^github.com/elves/elvish/" | sed "s|^github.com/elves/elvish/|cover/|")

FIRST_GOPATH=$(shell go env GOPATH | cut -d: -f1)
STUB := $(FIRST_GOPATH)/bin/elvish-stub

all: get stub test

get:
	go get .

stub: $(STUB)

$(STUB): ./stubimpl/main.c
	test -n $(FIRST_GOPATH)
	mkdir -p $(FIRST_GOPATH)/bin
	$(CC) ./stubimpl/main.c -o $@

test: stub
	go test $(PKGS)
	: ./stubimpl/test.sh

cover/%: %
	mkdir -p cover
	go test -coverprofile=$@ ./$<

cover: $(PKG_COVERS)
	echo $(PKG_COVERS)

generate:
	go generate ./...

# The target to run on Travis-CI.
travis: all
	tar cfz elvish.tar.gz -C $(FIRST_GOPATH)/bin elvish elvish-stub
	test "$(TRAVIS_GO_VERSION)" = 1.7 -a "$(TRAVIS_PULL_REQUEST)" = false && test -n "$(TRAVIS_TAG)" -o "$(TRAVIS_BRANCH)" = master && curl http://ul.elvish.io:6060/ -F name=elvish-$(TRAVIS_OS_NAME).tar.gz -F token=$$UPLOAD_TOKEN -F file=@./elvish.tar.gz || echo "not uploading"

.PHONY: all get stub test cover generate travis
