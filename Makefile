EXE := elvish
PKGS := edit eval parse util sys store errutil sysutil strutil
PKG_PATHS := $(addprefix ./,$(PKGS)) # go tools want an explicit ./
PKG_COVERS := $(addprefix cover/,$(PKGS))

all: elvish test

elvish:
	go get .

test:
	go test $(PKG_PATHS)

cover/%: %
	mkdir -p cover
	go test -coverprofile=$@ ./$<

cover: $(PKG_COVERS)

z-%.go: %.go
	go tool cgo -godefs $< > $@

pre-commit: sys/z-winsize.go

.PHONY: all elvish test cover pre-commit
