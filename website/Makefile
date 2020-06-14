TOOLS_DIR := ./tools
DST_DIR = ./_dst
PUBLISH_DIR := ./_publish

MDS := home.md $(filter-out %/README.md,$(wildcard [^_]*/*.md))
HTMLS := $(MDS:.md=.html)

TOOLS := highlight macros elvdoc genblog

default: gen

%.html: %.md $(TOOL_BINARIES) $(TOOLS_DIR)/md-to-html
	$(TOOLS_DIR)/md-to-html $< $@

# Generates the website into the dst directory.
gen: tools $(HTMLS)
	./$(TOOLS_DIR)/genblog.bin . $(DST_DIR)
	ln -sf `pwd`/favicons/* $(DST_DIR)/

tools:
	for tool in $(TOOLS); do \
		go build -o $(TOOLS_DIR)/$$tool.bin ./cmd/$$tool; \
	done

genblog:
	: # a nop, kept for backward compatibility with elves/up
	
# Synchronizes the generated website into $(PUBLISH_DIR). The PUBLISH_DIR
# environment variable can be overriden to a remote place to publish this online.
publish: gen
	rsync -aLv --delete ./_dst/ $(PUBLISH_DIR)/

clean:
	rm -f $(HTMLS) _tools/*.bin
	rm -rf $(DST_DIR)

.PHONY: default gen tools genblog publish clean
.SECONDARY: $(TOOL_BINARIES)
