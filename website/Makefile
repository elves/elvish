DST_DIR := ./_dst
PUBLISH_DIR := ./_publish
DOCSET_TMP_DIR := ./_docset_tmp
DOCSET_DST_DIR := ./Elvish.docset

MDS := home.md $(filter-out %/README.md,$(wildcard [^_]*/*.md))
HTMLS := $(MDS:.md=.html)

TOOL_PKGS := highlight macros elvdoc gensite
TOOL_BINS := $(addprefix tools/,$(addsuffix .bin,$(TOOL_PKGS)))

# Generates the website into $(DST_DIR).
gen: $(TOOL_BINS) $(HTMLS)
	tools/gensite.bin . $(DST_DIR)
	ln -sf `pwd`/fonts `pwd`/favicons/* $(DST_DIR)/

# Generates docset into $(DOCSET_DST_DIR).
docset: $(TOOL_BINS) $(HTMLS)
	ELVISH_DOCSET_MODE=1 tools/gensite.bin . $(DOCSET_TMP_DIR)
	./tools/mkdocset $(DOCSET_TMP_DIR) $(DOCSET_DST_DIR)

# Synchronizes the generated website into $(PUBLISH_DIR), which is passed to
# rsync and can be a remote place.
publish: gen
	rsync -aLv --delete ./_dst/ $(PUBLISH_DIR)/

check-rellinks: gen
	python3 tools/check-rellinks.py $(DST_DIR)

clean:
	rm -rf $(TOOL_BINS) $(HTMLS) $(DST_DIR) $(DOCSET_TMP_DIR) $(DOCSET_DST_DIR)

ttyshot/%.html:
	@echo Exclude ttyshots from the pattern rule for %.html

.PHONY: gen docset publish clean

.SECONDEXPANSION:

tools/%.bin: cmd/% $$(wildcard cmd/%/*.go)
	go build -o $@ ./$<

%.html: %.md $(TOOL_BINS) tools/md-to-html $$(shell tools/extra-deps $$@)
	tools/md-to-html $< $@
