// Package doc implements the doc: module.
package doc

import (
	"fmt"
	"io/fs"
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
	for ns, docs := range docs() {
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
			findIn(ns+entry.Name, entry.Content)
		}
		for _, entry := range docs.Vars {
			findIn("$"+ns+entry.Name, entry.Content)
		}
	}
}

// Source returns the doc source for a symbol.
func Source(fqname string) (string, error) {
	isVar := strings.HasPrefix(fqname, "$")
	if isVar {
		fqname = fqname[1:]
	}
	first, rest := eval.SplitQName(fqname)
	if rest == "" {
		first, rest = "", first
	} else if first == "builtin:" {
		first = ""
	}
	docs, ok := docs()[first]
	if !ok {
		return "", fmt.Errorf("no doc for %s", parse.Quote(fqname))
	}
	var entries []elvdoc.Entry
	if isVar {
		entries = docs.Vars
	} else {
		entries = docs.Fns
	}
	for _, entry := range entries {
		if entry.Name == rest {
			return entry.Content, nil
		}
	}

	return "", fmt.Errorf("no doc for %s", parse.Quote(fqname))
}

func symbols(fm *eval.Frame) error {
	var names []string
	for ns, docs := range docs() {
		for _, fn := range docs.Fns {
			names = append(names, ns+fn.Name)
		}
		for _, v := range docs.Vars {
			names = append(names, "$"+ns+v.Name)
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

var (
	docsOnce sync.Once
	docsMap  map[string]elvdoc.Docs
	// May be overridden in tests.
	elvFiles fs.FS = pkg.ElvFiles
)

// Returns a map from namespace prefixes (like "doc:", or "" for the builtin
// module) to extracted elvdocs.
//
// TODO: Simplify this using [sync.OnceValue] once Elvish requires Go 1.21.
func docs() map[string]elvdoc.Docs {
	docsOnce.Do(func() {
		// We don't expect any errors from reading an [embed.FS].
		docsMap, _ = elvdoc.ExtractAllFromFS(elvFiles)
	})
	return docsMap
}
