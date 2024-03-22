package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"src.elv.sh/pkg/elvdoc"
)

// Unlike Elvish's builtin documentation which embeds all the relevant .elv
// files into the binary itself, we read the filesystem at runtime. This allows
// us to read the new .elv file without rebuilding this program.
var pkgFS = os.DirFS("../pkg")

func filter(in io.Reader, out io.Writer) {
	f := filterer{}
	f.filter(in, out)
}

type filterer struct {
	module string
}

var macros = map[string]func(*filterer, string) string{
	"@module ": (*filterer).expandModule,
	"@dl ":     (*filterer).expandDl,
}

func (f *filterer) filter(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()
		for leader, expand := range macros {
			i := strings.Index(line, leader)
			if i >= 0 {
				line = line[:i] + expand(f, line[i+len(leader):])
				break
			}
		}
		fmt.Fprintln(out, line)
	}
	if f.module != "" {
		symbolPrefix := ""
		if f.module != "builtin" {
			symbolPrefix = f.module + ":"
		}
		docs, err := elvdoc.ExtractFromFS(pkgFS, symbolPrefix)
		if err != nil {
			log.Fatal(err)
		}
		var buf bytes.Buffer
		writeElvdocSections(&buf, docs)
		filter(&buf, out)
	}
}

func (f *filterer) expandModule(rest string) string {
	f.module = rest
	// Module doc will be added at end of file
	return fmt.Sprintf(
		"<a name='//apple_ref/cpp/Module/%s' class='dashAnchor'></a>", f.module)
}

func (f *filterer) expandDl(rest string) string {
	fields := strings.SplitN(rest, " ", 2)
	name := fields[0]
	url := name
	if len(fields) == 2 {
		url = fields[1]
	}
	return fmt.Sprintf(`<a href="https://dl.elv.sh/%s">%s</a>`, url, name)
}
