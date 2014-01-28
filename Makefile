EXE := das
PKGS := edit eval parse util
PKG_PATHS := $(addprefix ./,$(PKGS)) # go tools want an explicit ./
PKG_COVERAGES := $(addsuffix .coverage,$(PKGS))

exe:
	go build -o $(EXE) ./main

test:
	go test $(PKG_PATHS)

%.coverage: %
	go test -coverprofile=$@ ./$<

coverage: $(PKG_COVERAGES)

gofmt:
	gofmt -tabwidth=4 -w .

z-%.go: %.go
	go tool cgo -godefs $< > $@

pre-commit: gofmt edit/tty/z-types.go

.PHONY: exe test coverage gofmt pre-commit
