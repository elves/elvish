PKGS := $(filter-out main,$(shell go list -f '{{.Name}}' ./...))
PKG_COVERS := $(addprefix cover/,$(PKGS))

all: get test

get:
	go get .

test:
	go test ./...

cover/%: %
	mkdir -p cover
	go test -coverprofile=$@ ./$<

cover: $(PKG_COVERS)

generate:
	go generate ./...

# The target to run on Travis-CI.
travis: get test
	go build -o elvish-$(TRAVIS_OS_NAME)

.PHONY: all get test cover generate travis
