MDS := $(wildcard *.md)
HTMLS := $(MDS:.md=.html)

all: $(HTMLS)

%.html: %.md gen.elv template.html
	./gen.elv $< $@
