EXE := elvish
PKGS := edit eval parse util sys #service elvishd
PKG_PATHS := $(addprefix ./,$(PKGS)) # go tools want an explicit ./
PKG_COVERAGES := $(addprefix coverage/,$(PKGS))

all: elvish elvishd test

elvish:
	go get .

elvishd:
	go get ./elvishd

test:
	go test $(PKG_PATHS)

coverage/%: %
	mkdir -p coverage
	go test -coverprofile=$@ ./$<

coverage: $(PKG_COVERAGES)

z-%.go: %.go
	go tool cgo -godefs $< > $@

pre-commit: edit/tty/z-types.go

.PHONY: all elvish elvishd test coverage pre-commit
