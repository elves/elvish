package main

import (
	"fmt"
	"html"
	"io"
	"net/url"
	"sort"
	"strings"

	"src.elv.sh/pkg/elvdoc"
)

func writeElvdocSections(w io.Writer, ns string, docs elvdoc.Docs) {
	writeSection := func(heading, entryType, prefix string, entries []elvdoc.Entry) {
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
			attr := ""
			if entry.ID != "" {
				attr = " {#" + entry.ID + "}"
			}
			fmt.Fprintf(w, "## %s%s\n\n", fullName, attr)
			// The body is guaranteed to have a trailing newline, hence Fprint
			// instead of Fprintln.
			fmt.Fprint(w, entry.Content)
		}
	}

	if len(docs.Vars) > 0 {
		writeSection("Variables", "Variable", "$"+ns, docs.Vars)
	}
	if len(docs.Fns) > 0 {
		if len(docs.Vars) > 0 {
			fmt.Fprintln(w)
			fmt.Fprintln(w)
		}
		writeSection("Functions", "Function", ns, docs.Fns)
	}
}

var sortSymbol = map[string]string{
	"+": " a",
	"-": " b",
	"*": " c",
	"/": " d",
}

func symbolForSort(s string) string {
	// Hack to sort + - * / in that order, and before everything else.
	if t, ok := sortSymbol[strings.Fields(s)[0]]; ok {
		return t
	}
	// If there is a leading dash, move it to the end.
	if strings.HasPrefix(s, "-") {
		return s[1:] + "-"
	}
	return s
}
