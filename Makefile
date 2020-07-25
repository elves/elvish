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

test:
	echo "`go env GOOS`/`go env GOARCH`" | egrep -q '^(linux|freebsd|darwin|windows)/amd64$$' \
		&& go test -race ./... \
		|| go test ./...

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
.PHONY: default get generate test style checkstyle checkstyle-go checkstyle-md
