PKGS := $(filter-out main,$(shell go list -f '{{.Name}}' ./...))
PKG_COVERS := $(addprefix cover/,$(PKGS))

STUB := $(GOPATH)/bin/elvish-stub

all: get $(STUB) test

get:
	go get .

$(STUB): ./stubimpl/main.c
	test -n $(GOPATH)
	mkdir -p $(GOPATH)/bin
	$(CC) ./stubimpl/main.c -o $@

test: $(STUB)
	go test ./...
	: ./stubimpl/test.sh

cover/%: %
	mkdir -p cover
	go test -coverprofile=$@ ./$<

cover: $(PKG_COVERS)

generate:
	go generate ./...

# The target to run on Travis-CI.
travis: all
	tar cfJ elvish.tar.xz -C $(GOPATH)/bin elvish elvish-stub
	curl http://dl.elvish.io:6060/ -F name=elvish-$(if $(filter-out master,$(TRAVIS_BRANCH)),$(TRAVIS_BRANCH)-,)$(TRAVIS_OS_NAME).tar.xz -F token=$$UPLOAD_TOKEN -F file=@./elvish.tar.xz

.PHONY: all get test cover generate travis
