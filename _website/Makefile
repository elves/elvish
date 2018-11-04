TOOLS_DIR := ./_tools
DST_DIR = ./_dst
PUBLISH_DIR := ./_publish

MDS := home.md $(wildcard */*.md)
HTMLS := $(MDS:.md=.html)

TOOL_SOURCES := $(wildcard $(TOOLS_DIR)/*.go)
TOOL_BINARIES := $(TOOL_SOURCES:.go=.bin)

default: gen

%.html: %.md $(TOOL_BINARIES) $(TOOLS_DIR)/md-to-html
	$(TOOLS_DIR)/md-to-html $< $@

%.bin: %.go
	go build -o $@ ./$<

# Generates the website into the dst directory.
gen: $(HTMLS) maybe-genblog
	genblog . $(DST_DIR)
	rm -f $(DST_DIR)/assets && ln -sf `pwd`/_assets $(DST_DIR)/assets
	ln -sf `pwd`/_favicons/* $(DST_DIR)/

# Installs genblog only if not found.
maybe-genblog:
	which genblog || $(MAKE) genblog

# Installs latest genblog.
genblog:
	go get -u github.com/xiaq/genblog

# Copies the generated website into $(PUBLISH_DIR).
publish: gen
	rsync -aLv --delete ./_dst/ $(PUBLISH_DIR)/

clean:
	rm -f $(HTMLS) $(TOOL_BINARIES)
	rm -rf $(DST_DIR)

.PHONY: default gen maybe-genblog genblog publish clean
.SECONDARY: $(TOOL_BINARIES)
