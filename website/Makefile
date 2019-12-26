TOOLS_DIR := ./_tools
DST_DIR = ./_dst
PUBLISH_DIR := ./_publish

MDS := home.md $(wildcard [^_]*/*.md)
HTMLS := $(MDS:.md=.html)

TOOLS := highlight macros elvdoc

default: gen

%.html: %.md $(TOOL_BINARIES) $(TOOLS_DIR)/md-to-html
	$(TOOLS_DIR)/md-to-html $< $@

%.bin: %.go
	go build -o $@ ./$<

# Generates the website into the dst directory.
gen: tools $(HTMLS) maybe-genblog
	genblog . $(DST_DIR)
	ln -sf `pwd`/_favicons/* $(DST_DIR)/

tools:
	for tool in $(TOOLS); do \
		go build -o _tools/$$tool.bin ../cmd/website/$$tool; \
	done

# Installs genblog only if not found.
maybe-genblog:
	which genblog || $(MAKE) genblog

# Installs latest genblog.
genblog:
	go get -u github.com/xiaq/genblog

# Synchronizes the generated website into $(PUBLISH_DIR). The PUBLISH_DIR
# environment variable can be overriden to a remote place to publish this online.
publish: gen
	rsync -aLv --delete ./_dst/ $(PUBLISH_DIR)/

clean:
	rm -f $(HTMLS) _tools/*.bin
	rm -rf $(DST_DIR)

.PHONY: default gen tools maybe-genblog genblog publish clean
.SECONDARY: $(TOOL_BINARIES)
