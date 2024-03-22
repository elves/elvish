// Package doc implements the doc: module.
package doc

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"src.elv.sh/pkg"
	"src.elv.sh/pkg/elvdoc"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/md"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/sys"
)

var Ns = eval.BuildNsNamed("doc").
	AddGoFns(map[string]any{
		"show":     show,
		"find":     find,
		"source":   Source,
		"-symbols": symbols,
	}).
	Ns()

type showOptions struct{ Width int }

func (opts *showOptions) SetDefaultOptions() {}

func show(fm *eval.Frame, opts showOptions, fqname string) error {
	doc, err := Source(fqname)
	if err != nil {
		return err
	}
	width := opts.Width
	if width <= 0 {
		_, width = sys.WinSize(fm.Port(1).File)
		if width <= 0 {
			width = 80
		}
	}
	codec := &md.TTYCodec{
		Width:              width,
		HighlightCodeBlock: elvdoc.HighlightCodeBlock,
		ConvertRelativeLink: func(dest string) string {
			// TTYCodec does not show destinations of relative links by default.
			// Special-case links to language.html as they are quite common in
			// elvdocs.
			if strings.HasPrefix(dest, "language.html") {
				return "https://elv.sh/ref/" + dest
			}
			return ""
		},
	}
	if !strings.HasPrefix(fqname, "$") {
		// Function docs start with a code block that shows how to use the
		// function. Since the code block is indented 2 spaces by TTYCodec, it
		// looks a little bit weird as the first line of the output. Make the
		// output look slightly nicer by prepending a line.
		doc = "Usage:\n\n" + doc
	}
	_, err = fm.ByteOutput().WriteString(md.RenderString(doc, codec))
	return err
}

func find(fm *eval.Frame, qs ...string) {
	for _, docs := range docsMap() {
		findIn := func(name, markdown string) {
			if bs, ok := match(markdown, qs); ok {
				out := fm.ByteOutput()
				fmt.Fprintf(out, "%s:\n", name)
				for _, b := range bs {
					fmt.Fprintf(out, "  %s\n", b.Show())
				}
			}
		}
		for _, entry := range docs.Fns {
			findIn(entry.Name, entry.FullContent())
		}
		for _, entry := range docs.Vars {
			findIn(entry.Name, entry.FullContent())
		}
	}
}

// Source returns the doc source for a symbol.
func Source(qname string) (string, error) {
	isVar := strings.HasPrefix(qname, "$")
	var ns string
	if strings.ContainsRune(qname, ':') {
		if isVar {
			first, rest := eval.SplitQName(qname[1:])
			if first == "builtin:" {
				// Normalize $builtin:foo -> $foo, and leave ns as ""
				qname = "$" + rest
			} else {
				ns = first
			}
		} else {
			first, rest := eval.SplitQName(qname)
			if first == "builtin:" {
				// Normalize builtin:foo -> foo, and leave ns as ""
				qname = rest
			} else {
				ns = first
			}
		}
	}

	docs, ok := docsMap()[ns]
	if !ok {
		return "", fmt.Errorf("no doc for %s", parse.Quote(qname))
	}
	var entries []elvdoc.Entry
	if isVar {
		entries = docs.Vars
	} else {
		entries = docs.Fns
	}
	for _, entry := range entries {
		if entry.Name == qname {
			return entry.FullContent(), nil
		}
	}

	return "", fmt.Errorf("no doc for %s", parse.Quote(qname))
}

func symbols(fm *eval.Frame) error {
	var names []string
	for _, docs := range docsMap() {
		for _, fn := range docs.Fns {
			names = append(names, fn.Name)
		}
		for _, v := range docs.Vars {
			names = append(names, v.Name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		err := fm.ValueOutput().Put(name)
		if err != nil {
			return err
		}
	}
	return nil
}

// Can be overridden in tests.
var docsMapWithError = sync.OnceValues(func() (map[string]elvdoc.Docs, error) {
	return elvdoc.ExtractAllFromFS(pkg.ElvFiles)
})

// Returns a map from namespace prefixes (like "doc:", or "" for the builtin
// module) to extracted elvdocs.
func docsMap() map[string]elvdoc.Docs {
	// docsMapWithError depends on an embedded FS and returns the same values
	// every time. The error is checked in a test, so we don't have to check it
	// during runtime.
	m, _ := docsMapWithError()
	return m
}
