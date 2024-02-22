ELVISH_MAKE_BIN ?= $(or $(GOBIN),$(shell go env GOPATH)/bin)/elvish$(shell go env GOEXE)
ELVISH_MAKE_BIN := $(subst \,/,$(ELVISH_MAKE_BIN))
ELVISH_MAKE_PKG ?= ./cmd/elvish

default: test most-checks get

# This target emulates the behavior of "go install ./cmd/elvish", except that
# the build output and the main package to build can be overridden with
# environment variables.
get:
	mkdir -p $(shell dirname $(ELVISH_MAKE_BIN))
	go build -o $(ELVISH_MAKE_BIN) $(ELVISH_MAKE_PKG)

# Run formatters on Go and Markdown files.
fmt:
	find . -name '*.go' | xargs goimports -w
	find . -name '*.go' | xargs gofmt -s -w
	find . -name '*.md' | xargs go run src.elv.sh/cmd/elvmdfmt -w -width 80

# Run unit tests, with race detection if the platform supports it.
test:
	go test $(shell ./tools/run-race.elv) ./...
	cd website; go test $(shell ./tools/run-race.elv) ./...

# Generate a basic test coverage report, and open it in the browser. The report
# is an approximation of https://app.codecov.io/gh/elves/elvish/.
cover:
	go test -coverprofile=cover -coverpkg=./pkg/... ./pkg/...
	./tools/prune-cover.sh .codecov.yml cover
	go tool cover -html=cover
	go tool cover -func=cover | tail -1 | awk '{ print "Overall coverage:", $$NF }'

# All the checks except check-gen.sh, which is not always convenient to run as
# it requires a clean working tree.
most-checks:
	./tools/check-fmt-go.sh
	./tools/check-fmt-md.sh
	./tools/check-disallowed.sh
	codespell
	go vet ./...
	staticcheck ./...

all-checks: most-checks
	./tools/check-gen.sh

.PHONY: default get fmt test cover most-checks all-checks
