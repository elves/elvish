PKGS := edit errutil eval logutil osutil parse print store strutil sys
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

.PHONY: all get test cover generate
