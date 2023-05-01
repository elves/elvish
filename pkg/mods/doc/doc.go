// Package doc implements the doc: module.
package doc

import (
	"embed"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	"src.elv.sh/pkg/edit"
	"src.elv.sh/pkg/elvdoc"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/md"
	"src.elv.sh/pkg/mods/epm"
	"src.elv.sh/pkg/mods/file"
	"src.elv.sh/pkg/mods/flag"
	"src.elv.sh/pkg/mods/math"
	"src.elv.sh/pkg/mods/path"
	"src.elv.sh/pkg/mods/platform"
	"src.elv.sh/pkg/mods/re"
	"src.elv.sh/pkg/mods/readlinebinding"
	"src.elv.sh/pkg/mods/runtime"
	"src.elv.sh/pkg/mods/store"
	"src.elv.sh/pkg/mods/str"
	"src.elv.sh/pkg/mods/unix"
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

// DElvCode contains the content of the .d.elv file for this module.
//
//go:embed *.d.elv
var DElvCode string

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
		HighlightCodeBlock: HighlightCodeBlock,
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
	for ns, docs := range Docs() {
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
	docs, ok := Docs()[first]
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
	for ns, docs := range Docs() {
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

var modToCode = map[string]io.Reader{
	"":                 readAll(eval.BuiltinDElvFiles),
	"doc:":             read(DElvCode),
	"edit:":            readAll(edit.DElvFiles),
	"epm:":             read(epm.Code),
	"file:":            read(file.DElvCode),
	"flag:":            read(flag.DElvCode),
	"math:":            read(math.DElvCode),
	"path:":            read(path.DElvCode),
	"platform:":        read(platform.DElvCode),
	"re:":              read(re.DElvCode),
	"readlinebinding:": read(readlinebinding.Code),
	"runtime:":         read(runtime.DElvCode),
	"store:":           read(store.DElvCode),
	"str:":             read(str.DElvCode),
	"unix:":            readAll(unix.DElvFiles),
}

var (
	docsOnce sync.Once
	docs     map[string]elvdoc.Docs
)

// Docs returns a map from namespace prefixes (like "doc:", or "" for the
// builtin module) to extracted elvdocs.
func Docs() map[string]elvdoc.Docs {
	docsOnce.Do(func() {
		docs = make(map[string]elvdoc.Docs, len(modToCode))
		for mod, code := range modToCode {
			docs[mod], _ = elvdoc.Extract(code, mod)
		}
	})
	return docs
}

func read(s string) io.Reader { return strings.NewReader(s) }

func readAll(fs embed.FS) io.Reader {
	entries, _ := fs.ReadDir(".")
	readers := make([]io.Reader, 2*len(entries)-1)
	for i, entry := range entries {
		readers[2*i], _ = fs.Open(entry.Name())
		if i < len(entries)-1 {
			// Insert an empty line between adjacent files so that the comment
			// block at the end of one file doesn't get merged with the comment
			// block at the start of the next file.
			readers[2*i+1] = strings.NewReader("\n\n")
		}
	}
	return io.MultiReader(readers...)
}
