default: test get

get:
	go get -trimpath -ldflags \
		"-X github.com/elves/elvish/pkg/buildinfo.Version=$$(git describe --tags --always --dirty=-dirty) \
		 -X github.com/elves/elvish/pkg/buildinfo.Reproducible=true" .

# Used by elves/up
buildall:
	./tools/buildall.sh

generate:
	go generate ./...

# Run unit tests -- with race detection if the platform supports it.
test:
	if echo `go env GOOS GOARCH` | egrep -qx '(linux|freebsd|darwin|windows) amd64'; then \
		go test -race ./... ; \
	else \
		go test ./... ; \
	fi

# Generate a basic test coverage report. This will open the report in your
# browser. See also https://codecov.io/gh/elves/elvish/.
cover:
	rm -f elvish-coverage.dat
	go test -covermode=set -coverprofile=./elvish-coverage.dat ./...
	go tool cover -html=./elvish-coverage.dat

# Ensure the style of Go and Markdown source files is consistent.
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

.SILENT: checkstyle-go checkstyle-md
.PHONY: default get generate test style checkstyle checkstyle-go checkstyle-md cover
