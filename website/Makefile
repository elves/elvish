PUBLISH_DIR := ./published # ul.elvish.io:elvish.io/published

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
gen: $(HTMLS)
	genblog src dst

# Installs or updates genblog.
genblog:
	cd $(GOPATH)/src/github.com/xiaq/genblog; \
		git pull; \
		go generate; \
		go get
	genblog -print-default-css > assets/genblog.css

# Copies the generated website into $(PUBLISH_DIR).
publish: gen
	rsync -aLv --delete ./dst/ $(PUBLISH_DIR)/

clean:
	rm -f $(HTMLS) $(TOOL_BINARIES)

.PHONY: default gen genblog publish clean
.SECONDARY: $(TOOL_BINARIES)
