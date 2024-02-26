DST_DIR := ./_dst
PUBLISH_DIR := ./_publish
DOCSET_TMP_DIR := ./_docset_tmp
DOCSET_DST_DIR := ./Elvish.docset

MDS := home.md $(filter-out %/README.md,$(wildcard [^_]*/*.md))
HTMLS := $(MDS:.md=.html)

# Generates the website into $(DST_DIR).
gen: tools/gensite.bin $(HTMLS)
	tools/gensite.bin . $(DST_DIR)
	ln -sf `pwd`/fonts `pwd`/favicons/* $(DST_DIR)/

# Generates docset into $(DOCSET_DST_DIR).
docset: tools/gensite.bin $(HTMLS)
	ELVISH_DOCSET_MODE=1 tools/gensite.bin . $(DOCSET_TMP_DIR)
	tools/mkdocset $(DOCSET_TMP_DIR) $(DOCSET_DST_DIR)

# Synchronizes the generated website into $(PUBLISH_DIR), which is passed to
# rsync and can be a remote place.
publish: gen
	rsync -aLv --delete ./_dst/ $(PUBLISH_DIR)/

check-rellinks: gen
	python3 tools/check-rellinks.py $(DST_DIR)

clean:
	rm -rf tools/*.bin $(HTMLS) $(DST_DIR) $(DOCSET_TMP_DIR) $(DOCSET_DST_DIR)

ifdef TTYSHOT
%-ttyshot.html: %-ttyshot.elvts tools/ttyshot.bin
	tools/ttyshot.bin $(if $(findstring verbose,$(TTYSHOT)),-v,) -o $@ $<
else
%-ttyshot.html:
	@: ttyshot generation disabled by default
endif

.PHONY: gen docset publish check-rellinks clean

# Don't remove intermediate targets
.SECONDARY:

# Rules below have dynamic prerequisite lists, which requires GNU Make's
# .SECONDEXPANSION.
.SECONDEXPANSION:

tools/%.bin: cmd/% $$(wildcard cmd/%/*) go.mod ../go.mod $$(shell tools/cmd-deps ./cmd/%)
	go build -o $@ ./$<

%.html: %.md tools/md2html.bin $$(shell tools/md-deps $$@)
	tools/md2html.bin < $< > $@
