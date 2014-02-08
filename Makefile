EXE := elvish
PKGS := edit eval parse util
PKG_PATHS := $(addprefix ./,$(PKGS)) # go tools want an explicit ./
PKG_COVERAGES := $(addprefix coverage/,$(PKGS))

all: bin test

bin:
	go install .

test:
	go test $(PKG_PATHS)

coverage/%: %
	mkdir -p coverage
	go test -coverprofile=$@ ./$<

coverage: $(PKG_COVERAGES)

gofmt:
	gofmt -tabwidth=4 -w .

z-%.go: %.go
	go tool cgo -godefs $< > $@

pre-commit: gofmt edit/tty/z-types.go

.PHONY: all bin test coverage gofmt pre-commit
