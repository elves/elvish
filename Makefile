PKG_BASE := github.com/elves/elvish
PKGS := $(shell go list ./... | sed 's|^$(PKG_BASE)|.|')
PKG_COVERS := $(shell go list ./... | sed 's|^$(PKG_BASE)|.|' | grep -v '^\.$$' | sed 's/^\./cover/' | sed 's/$$/.cover/')
COVER_MODE := set
VERSION := $(shell git describe --tags --always --dirty=-dirty)

# -race requires cgo
ifneq ($(OS),Windows_NT)
    TEST_ENV := CGO_ENABLED=1
    TEST_FLAGS := -race
endif

default: test get

get:
	go get -trimpath -ldflags \
		"-X github.com/elves/elvish/pkg/buildinfo.Version=$(VERSION) \
		 -X github.com/elves/elvish/pkg/buildinfo.Reproducible=true" .

# Used by elves/up
buildall:
	./tools/buildall.sh

generate:
	go generate ./...

test:
	$(TEST_ENV) go test $(TEST_FLAGS) $(PKGS)

style:
	find . -name '*.go' | xargs goimports -w
	find . -name '*.md' | xargs prettier --tab-width 4 --prose-wrap always --write

checkstyle: checkstyle-go checkstyle-md

checkstyle-go:
	echo 'Go files that need formatting:'
	! find . -name '*.go' | xargs goimports -l \
		| sed 's/^/  /' | grep . && echo '  None!'

checkstyle-md:
	echo 'Markdown files that need formatting:'
	! find . -name '*.md' | xargs prettier --tab-width 4 --prose-wrap always -l \
		| sed 's/^/  /' | grep . && echo '  None!'

cover/%.cover: %
	mkdir -p $(dir $@)
	go test -coverprofile=$@ -covermode=$(COVER_MODE) ./$<

cover/all: $(PKG_COVERS)
	echo mode: $(COVER_MODE) > $@
	for f in $(PKG_COVERS); do test -f $$f && sed 1d $$f >> $@ || true; done

.SILENT: checkstyle-go checkstyle-md
.PHONY: default get generate test style checkstyle checkstyle-go checkstyle-md
