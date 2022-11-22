package main

import (
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"src.elv.sh/pkg/elvdoc"
)

func main() {
	run(os.Args[1:], os.Stdin, os.Stdout)
}

func run(args []string, in io.Reader, out io.Writer) {
	flags := flag.NewFlagSet("", flag.ExitOnError)
	var ns = flags.String("ns", "", "namespace prefix")

	err := flags.Parse(args)
	if err != nil {
		log.Fatal(err)
	}
	args = flags.Args()

	if len(args) > 0 {
		extractPaths(args, *ns, out)
	} else {
		extract(in, *ns, out)
	}
}

func extractPaths(paths []string, ns string, out io.Writer) {
	var files []string
	for _, path := range paths {
		stat, err := os.Stat(path)
		if err != nil {
			log.Fatal(err)
		}
		if stat.IsDir() {
			elvFiles := mustGlob(filepath.Join(path, "*.elv"))
			files = append(files, elvFiles...)
		} else {
			files = append(files, path)
		}
	}

	reader, cleanup, err := multiFile(files)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()
	extract(reader, ns, out)
}

func mustGlob(p string) []string {
	files, err := filepath.Glob(p)
	if err != nil {
		log.Fatal(err)
	}
	return files
}

// Makes a reader that concatenates multiple files.
func multiFile(names []string) (io.Reader, func(), error) {
	readers := make([]io.Reader, 2*len(names))
	closers := make([]io.Closer, len(names))
	for i, name := range names {
		file, err := os.Open(name)
		if err != nil {
			for j := 0; j < i; j++ {
				closers[j].Close()
			}
			return nil, nil, err
		}
		readers[2*i] = file
		// Insert an empty line between adjacent files so that the comment block
		// at the end of one file doesn't get merged with the comment block at
		// the start of the next file.
		readers[2*i+1] = strings.NewReader("\n\n")
		closers[i] = file
	}
	return io.MultiReader(readers...), func() {
		for _, closer := range closers {
			closer.Close()
		}
	}, nil
}

func extract(r io.Reader, ns string, w io.Writer) {
	docs, err := elvdoc.Extract(r, ns)
	if err != nil {
		log.Fatal(err)
	}

	write := func(heading, entryType, prefix string, entries []elvdoc.Entry) {
		fmt.Fprintf(w, "# %s\n", heading)
		sort.Slice(entries, func(i, j int) bool {
			return symbolForSort(entries[i].Name) < symbolForSort(entries[j].Name)
		})
		for _, entry := range entries {
			fmt.Fprintln(w)
			fullName := prefix + entry.Name
			// Create anchors for Docset. These anchors are used to show a ToC;
			// the mkdsidx.py script also looks for those anchors to generate
			// the SQLite index.
			//
			// Some builtin commands are documented together. Create an anchor
			// for each of them.
			for _, s := range strings.Fields(fullName) {
				fmt.Fprintf(w,
					"<a name='//apple_ref/cpp/%s/%s' class='dashAnchor'></a>\n\n",
					entryType, url.QueryEscape(html.UnescapeString(s)))
			}
			id := entry.ID
			if id == "" {
				// pandoc strips punctuations from the ID, turning "mod:name"
				// into "modname". Explicitly preserve the original full name
				// by specifying an attribute. We still strip the leading $ for
				// variables since pandoc will treat "{#$foo}" as part of the
				// title.
				id = strings.TrimPrefix(fullName, "$")
			}
			fmt.Fprintf(w, "## %s {#%s}\n\n", fullName, id)
			// The body is guaranteed to have a trailing newline, hence Fprint
			// instead of Fprintln.
			fmt.Fprint(w, entry.Content)
		}
	}

	if len(docs.Vars) > 0 {
		write("Variables", "Variable", "$"+ns, docs.Vars)
	}
	if len(docs.Fns) > 0 {
		if len(docs.Vars) > 0 {
			fmt.Fprintln(w)
			fmt.Fprintln(w)
		}
		write("Functions", "Function", ns, docs.Fns)
	}
}

var sortSymbol = map[string]string{
	"+": " a",
	"-": " b",
	"*": " c",
	"/": " d",
}

func symbolForSort(s string) string {
	// Hack to sort + - * / in that order.
	if t, ok := sortSymbol[strings.Fields(s)[0]]; ok {
		return t
	}
	// If there is a leading dash, move it to the end.
	if strings.HasPrefix(s, "-") {
		return s[1:] + "-"
	}
	return s
}
