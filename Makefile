ELVISH_MAKE_BIN ?= $(or $(GOBIN),$(shell go env GOPATH)/bin)/elvish$(shell go env GOEXE)
ELVISH_MAKE_BIN := $(subst \,/,$(ELVISH_MAKE_BIN))
ELVISH_MAKE_PKG ?= ./cmd/elvish

default: test get

# This target emulates the behavior of "go install ./cmd/elvish", except that
# the build output and the main package to build can be overridden with
# environment variables.
get:
	mkdir -p $(shell dirname $(ELVISH_MAKE_BIN))
	go build -o $(ELVISH_MAKE_BIN) $(ELVISH_MAKE_PKG)

generate:
	go generate ./...

# Run unit tests, with race detection if the platform supports it.
test:
	go test $(shell ./tools/run-race.sh) ./...
	cd website; go test $(shell ./tools/run-race.sh) ./...

# Generate a basic test coverage report, and open it in the browser. See also
# https://apps.codecov.io/gh/elves/elvish/.
cover:
	go test -coverprofile=cover -coverpkg=./pkg/... ./pkg/...
	./tools/prune-cover.sh .codecov.yml cover
	go tool cover -html=cover
	go tool cover -func=cover | tail -1 | awk '{ print "Overall coverage:", $$NF }'

# Ensure the style of Go and Markdown source files is consistent.
style:
	find . -name '*.go' | xargs goimports -w
	find . -name '*.go' | xargs gofmt -s -w
	find . -name '*.md' | xargs go run src.elv.sh/cmd/elvmdfmt -w -width 80

# Check if the style of the Go and Markdown files is correct without modifying
# those files.
checkstyle: checkstyle-go checkstyle-md

checkstyle-go:
	./tools/checkstyle-go.sh

checkstyle-md:
	./tools/checkstyle-md.sh

lint:
	./tools/lint.sh

codespell:
	codespell --skip .git

check-disallowed:
	./tools/check-disallowed.sh

check-gen:
	./tools/check-gen.sh

.SILENT: checkstyle-go checkstyle-md lint
.PHONY: default get generate test cover style checkstyle checkstyle-go checkstyle-md lint codespell check-disallowed check-gen
