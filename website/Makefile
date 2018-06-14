PUBLISH_DIR := ./publish

MDS := $(wildcard src/*.md src/*/*.md)
HTMLS := $(MDS:.md=.html)

TOOL_SOURCES := $(wildcard *.go)
TOOL_BINARIES := $(TOOL_SOURCES:.go=.tool)

default: gen

%.html: %.md $(TOOL_BINARIES) md-to-html
	./md-to-html $< $@

%.tool: %.go
	go build -o $@ $<

# Generates the website into the dst directory.
gen: $(HTMLS) maybe-genblog
	genblog src dst
	ln -sf `pwd`/assets dst/
	ln -sf `pwd`/favicons/* dst/

# Installs genblog only if not found.
maybe-genblog:
	which genblog || make genblog

# Installs latest genblog.
genblog:
	go get -u github.com/xiaq/genblog
	genblog -print-default-css > assets/genblog.css

# Copies the generated website into $(PUBLISH_DIR).
publish: gen
	rsync -aLv --delete ./dst/ $(PUBLISH_DIR)/

clean:
	rm -f $(HTMLS) $(TOOL_BINARIES)

.PHONY: default gen maybe-genblog genblog publish clean
.SECONDARY: $(TOOL_BINARIES)
