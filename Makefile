PKGS := $(filter-out main,$(shell go list -f '{{.Name}}' ./...))
PKG_COVERS := $(addprefix cover/,$(PKGS))

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
	tar cfz elvish.tar.gz -C $(GOPATH)/bin elvish elvish-stub
	curl http://athens.xiaq.me:6060/ -F name=elvish-$(if $(filter-out master,$(TRAVIS_BRANCH)),$(TRAVIS_BRANCH)-,)$(TRAVIS_OS_NAME).tar.gz -F token=$$UPLOAD_TOKEN -F file=@./elvish.tar.gz
	# curl http://ul.elvish.io:6060/ -F name=elvish-$(if $(filter-out master,$(TRAVIS_BRANCH)),$(TRAVIS_BRANCH)-,)$(TRAVIS_OS_NAME).tar.gz -F token=$$UPLOAD_TOKEN -F file=@./elvish.tar.gz

.PHONY: all get stub test cover generate travis
