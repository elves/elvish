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

func writeElvdocSections(w io.Writer, docs elvdoc.Docs) {
	writeSection := func(heading, entryType string, entries []elvdoc.Entry) {
		fmt.Fprintf(w, "# %s\n", heading)
		sort.Slice(entries, func(i, j int) bool {
			return symbolForSort(entries[i].Name) < symbolForSort(entries[j].Name)
		})
		for _, entry := range entries {
			fmt.Fprintln(w)
			// Create anchors for Docset. These anchors are used to show a ToC;
			// the mkdsidx.py script also looks for those anchors to generate
			// the SQLite index.
			//
			// Some builtin commands are documented together. Create an anchor
			// for each of them.
			for _, s := range strings.Fields(entry.Name) {
				fmt.Fprintf(w,
					"<a name='//apple_ref/cpp/%s/%s' class='dashAnchor'></a>\n\n",
					entryType, url.QueryEscape(html.UnescapeString(s)))
			}
			attr := ""
			for _, directive := range entry.Directives {
				if htmlID, ok := strings.CutPrefix(directive, "doc:html-id "); ok {
					attr = " {#" + strings.TrimSpace(htmlID) + "}"
				}
			}
			fmt.Fprintf(w, "## %s%s\n\n", entry.Name, attr)
			// The body is guaranteed to have a trailing newline, hence Fprint
			// instead of Fprintln.
			fmt.Fprint(w, entry.FullContent())
		}
	}

	if len(docs.Vars) > 0 {
		writeSection("Variables", "Variable", docs.Vars)
	}
	if len(docs.Fns) > 0 {
		if len(docs.Vars) > 0 {
			fmt.Fprintln(w)
			fmt.Fprintln(w)
		}
		writeSection("Functions", "Function", docs.Fns)
	}
}

func symbolForSort(s string) string {
	// Hack to sort unstable symbols close to their stable counterparts: for
	// example, let "-gc" appear between "gb" and "gd", but after "gc".
	if strings.HasPrefix(s, "-") {
		return s[1:] + "-"
	}
	return s
}
