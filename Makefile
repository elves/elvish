EXE := elvish
PKGS := edit eval parse util sys #service elvishd
PKG_PATHS := $(addprefix ./,$(PKGS)) # go tools want an explicit ./
PKG_COVERS := $(addprefix cover/,$(PKGS))

all: elvish elvishd test

elvish:
	go get .

elvishd:
	go get ./elvishd

test:
	go test $(PKG_PATHS)

cover/%: %
	mkdir -p cover
	go test -coverprofile=$@ ./$<

cover: $(PKG_COVERS)

z-%.go: %.go
	go tool cgo -godefs $< > $@

pre-commit: edit/tty/z-types.go

.PHONY: all elvish elvishd test cover pre-commit
