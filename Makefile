PKGS := $(shell go list ./... | grep -v /vendor/)
PKG_COVERS := $(shell go list ./... | grep -v /vendor/ | grep "^github.com/elves/elvish/" | sed "s|^github.com/elves/elvish/|cover/|")
COVER_MODE := count

FIRST_GOPATH=$(shell go env GOPATH | cut -d: -f1)
STUB := $(FIRST_GOPATH)/bin/elvish-stub

default: get stub test

get:
	go get .

generate:
	go generate ./...

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
	go test -coverprofile=$@ -covermode=$(COVER_MODE) ./$<

cover/all: $(PKG_COVERS)
	echo mode: $(COVER_MODE) > $@
	for f in $(PKG_COVERS); do test -f $$f && sed 1d $$f >> $@; done

goveralls: cover/all
	go get -u github.com/mattn/goveralls
	$(FIRST_GOPATH)/bin/goveralls -coverprofile=cover/all -service=travis-ci \

upload: get stub
	tar cfz elvish.tar.gz -C $(FIRST_GOPATH)/bin elvish elvish-stub
	test "$(TRAVIS_GO_VERSION)" = 1.7 -a "$(TRAVIS_PULL_REQUEST)" = false \
		&& test -n "$(TRAVIS_TAG)" -o "$(TRAVIS_BRANCH)" = master \
		&& curl http://ul.elvish.io:6060/ -F name=elvish-$(TRAVIS_OS_NAME).tar.gz \
			-F token=$$UPLOAD_TOKEN -F file=@./elvish.tar.gz\
		|| echo "not uploading"

travis: goveralls upload

.PHONY: default get generate stub test goveralls upload travis
