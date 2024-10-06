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
	mkdir -p _cover/unit _cover/e2e
	# Generate coverage from unit tests. We could generate text profiles
	# directly with -coverprofile, but there's no support for merging multiple
	# text profiles. So we generate binary profiles instead.
	go test -coverpkg=./pkg/... ./pkg/... -test.gocoverdir $$PWD/_cover/unit
	# Generate coverage from E2E tests, using -count to skip the cache.
	env GOCOVERDIR=$$PWD/_cover/e2e go test -count 1 ./e2e
	# Merge and convert binary profiles to a single text profile.
	go tool covdata textfmt -i _cover/unit,_cover/e2e -o _cover/merged.txt
	./tools/prune-cover.sh .codecov.yml < _cover/merged.txt > _cover/pruned.txt
	go tool cover -html _cover/pruned.txt
	go tool cover -func _cover/pruned.txt | tail -1 | awk '{ print "Overall coverage:", $$NF }'

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
