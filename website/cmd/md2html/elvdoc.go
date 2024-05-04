package main

import (
	"fmt"
	"html"
	"io"
	"net/url"
	"os"
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
			// Convert directives into header attributes
			// (https://pandoc.org/MANUAL.html#extension-header_attributes),
			// which will get interpreted by [htmlCodec.Do].
			var attrs []string
			for _, directive := range entry.Directives {
				if htmlID, ok := strings.CutPrefix(directive, "doc:html-id "); ok {
					attrs = append(attrs, "#"+strings.TrimSpace(htmlID))
				} else if addedIn, ok := strings.CutPrefix(directive, "doc:added-in "); ok {
					attrs = append(attrs, "added-in="+addedIn)
				} else if strings.HasPrefix(directive, "doc:") {
					fmt.Fprintf(os.Stderr, "\033[31mWarning: unknown directive: %s\033[m\n", directive)
				}
			}
			attrString := ""
			if len(attrs) > 0 {
				attrString = " {" + strings.Join(attrs, " ") + "}"
			}

			// Print the header.
			fmt.Fprintf(w, "## %s%s\n\n", entry.Name, attrString)
			// Print the body - it's is guaranteed to have a trailing newline,
			// hence Fprint instead of Fprintln.
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
