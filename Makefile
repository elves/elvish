PKGS := $(shell go list ./... | grep -v /vendor/)
PKG_COVERS := $(shell go list ./... | grep -v /vendor/ | grep "^github.com/elves/elvish/" | sed "s|^github.com/elves/elvish/|cover/|")

STUB := $(GOPATH)/bin/elvish-stub

all: get stub test

get:
	go get .

stub: $(STUB)

$(STUB): ./stubimpl/main.c
	test -n $(GOPATH)
	mkdir -p $(GOPATH)/bin
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
	tar cfz elvish.tar.gz -C $(GOPATH)/bin elvish elvish-stub
	test "$(TRAVIS_GO_VERSION)" == 1.7 && curl http://ul.elvish.io:6060/ -F name=elvish-$(if $(filter-out master,$(TRAVIS_BRANCH)),$(TRAVIS_BRANCH)-,)$(TRAVIS_OS_NAME).tar.gz -F token=$$UPLOAD_TOKEN -F file=@./elvish.tar.gz || true

.PHONY: all get stub test cover generate travis
