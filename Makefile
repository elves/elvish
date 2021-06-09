ELVISH_MAKE_BIN ?= $(shell go env GOPATH)/bin/elvish
ifdef ELVISH_MAKE_TAGS
	override ELVISH_MAKE_TAGS := -tags ${ELVISH_MAKE_TAGS}
endif

default: test get

get:
	export CGO_ENABLED=0; \
	if go env GOOS GOARCH | egrep -qx '(windows .*|linux (amd64|arm64))'; then \
		export GOFLAGS=-buildmode=pie; \
	fi; \
	mkdir -p $(shell dirname $(ELVISH_MAKE_BIN))
	go build -o $(ELVISH_MAKE_BIN) $(ELVISH_MAKE_TAGS) -trimpath -ldflags \
		"-X src.elv.sh/pkg/buildinfo.VersionSuffix=-dev.$$(git rev-parse HEAD)$$(git diff HEAD --quiet || printf +%s `uname -n`) \
		 -X src.elv.sh/pkg/buildinfo.Reproducible=true" ./cmd/elvish

generate:
	go generate ./...

# Run unit tests, with race detection if the platform supports it.
test:
	go test $(shell ./tools/run-race.sh) $(ELVISH_MAKE_TAGS) ./...

# Generate a basic test coverage report, and open it in the browser. See also
# https://apps.codecov.io/gh/elves/elvish/.
cover:
	go test -coverprofile=cover -coverpkg=./pkg/... $(ELVISH_MAKE_TAGS) ./pkg/...
	go tool cover -html=cover
	go tool cover -func=cover | tail -1 | awk '{ print "Overall coverage:", $$NF }'

# Ensure the style of Go and Markdown source files is consistent.
style:
	find . -name '*.go' | xargs goimports -w
	find . -name '*.go' | xargs gofmt -s -w
	find . -name '*.md' | xargs prettier --tab-width 4 --prose-wrap always --write

# Check if the style of the Go and Markdown files is correct without modifying
# those files.
checkstyle: checkstyle-go checkstyle-md

checkstyle-go:
	./tools/checkstyle-go.sh

checkstyle-md:
	./tools/checkstyle-md.sh

.SILENT: checkstyle-go checkstyle-md
.PHONY: default get generate test style checkstyle checkstyle-go checkstyle-md cover
