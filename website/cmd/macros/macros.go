// The macros program implements an ad-hoc preprocessor for Markdown files, used
// in Elvish's website.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

const (
	ttyshot = "@ttyshot "
	cf      = "@cf "
	dl      = "@dl "
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		line = expandTtyshot(line)
		line = expandCf(line)
		line = expandDl(line)
		fmt.Println(line)
	}
}

func expandTtyshot(line string) string {
	i := strings.Index(line, ttyshot)
	if i < 0 {
		return line
	}
	name := line[i+len(ttyshot):]
	content, err := ioutil.ReadFile(path.Join("ttyshot", name+".html"))
	if err != nil {
		log.Fatal(err)
	}
	var buf bytes.Buffer
	buf.WriteString(line[:i])
	buf.WriteString(`<pre class="ttyshot"><code>`)
	buf.Write(bytes.Replace(
		content, []byte("\n"), []byte("<br>"), -1))
	buf.WriteString("</code></pre>")
	return buf.String()
}

func expandCf(line string) string {
	i := strings.Index(line, cf)
	if i < 0 {
		return line
	}
	targets := strings.Split(line[i+len(cf):], " ")
	var buf strings.Builder
	buf.WriteString("See also")
	for i, target := range targets {
		if i == 0 {
			buf.WriteString(" ")
		} else if i == len(targets)-1 {
			buf.WriteString(" and ")
		} else {
			buf.WriteString(", ")
		}
		fmt.Fprintf(&buf, "[`%s`](%s)", target, cfHref(target))
	}
	buf.WriteString(".")
	return buf.String()
}

// Returns the href for a `@cf` reference.
func cfHref(target string) string {
	i := strings.IndexRune(target, ':')
	if i == -1 {
		// A link within the builtin page. Use unqualified name (e.g. #put).
		return "#" + target
	}

	module, symbol := target[:i], target[i+1:]
	if module == "builtin" {
		// A link from outside the builtin page to the builtin page. Use
		// unqualified name (e.g. #put).
		return "builtin.html#" + symbol
	}
	// A link to a non-builtin page. The section names are always qualified
	// names (e.g str:join), but pandoc strips the colons from the anchor (e.g.
	// #strjoin).
	return module + ".html#" + strings.ReplaceAll(target, ":", "")
}

func expandDl(line string) string {
	i := strings.Index(line, dl)
	if i < 0 {
		return line
	}
	fields := strings.SplitN(line[i+len(dl):], " ", 2)
	name := fields[0]
	url := name
	if len(fields) == 2 {
		url = fields[1]
	}
	return line[:i] + fmt.Sprintf(
		`<a href="https://dl.elv.sh/%s">%s</a>`, url, name)
}
