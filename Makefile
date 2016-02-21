PKGS := $(filter-out main,$(shell go list -f '{{.Name}}' ./...))
PKG_COVERS := $(addprefix cover/,$(PKGS))

all: get test

get:
	go get .
	cc ./stubimpl/main.c -o $(GOPATH)/bin/elvish-stub

test:
	go test ./...
	: ./stubimpl/test.sh

cover/%: %
	mkdir -p cover
	go test -coverprofile=$@ ./$<

cover: $(PKG_COVERS)

generate:
	go generate ./...

# The target to run on Travis-CI.
travis: get test
	tar cfJ elvish.tar.xz -C $(GOPATH)/bin elvish elvish-stub
	curl http://dl.elvish.io:6060/ -F name=elvish-$(TRAVIS_OS_NAME).tar.xz -F token=$$UPLOAD_TOKEN -F file=@./elvish.tar.xz

.PHONY: all get test cover generate travis
