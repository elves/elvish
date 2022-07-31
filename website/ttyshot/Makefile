SPECS := $(shell find . -name '*.spec')
HTMLS := $(SPECS:.spec=.html)

all: $(HTMLS)

%.html: %.spec ../tools/ttyshot.bin
	../tools/ttyshot.bin $<

.PHONY: all

.SECONDEXPANSION:

../tools/ttyshot.bin: $$(wildcard ../cmd/ttyshot/*)
	make -C .. tools/ttyshot.bin
