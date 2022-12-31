// Package doc implements the doc: module.
package doc

import (
	"embed"
	"fmt"
	"io"
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
		"show":   show,
		"source": source,
	}).
	Ns()

// DElvCode contains the content of the .d.elv file for this module.
//
//go:embed *.d.elv
var DElvCode string

type showOptions struct{ Width int }

func (opts *showOptions) SetDefaultOptions() {}

func show(fm *eval.Frame, opts showOptions, fqname string) error {
	doc, err := source(fqname)
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
	_, err = fm.ByteOutput().WriteString(md.RenderString(doc, &md.TTYCodec{Width: width}))
	return err
}

func source(fqname string) (string, error) {
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
