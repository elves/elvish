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
	content, err := ioutil.ReadFile(path.Join("_ttyshot", name+".html"))
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

// Given a `@cf` reference convert it to a HTML link.
func makeLink(cf string) string {
	i := strings.IndexRune(cf, ':')
	if i == -1 {
		// This handles the @cf uses in the `builtin:` namespace to other
		// sections in the same namespace.
		return fmt.Sprintf("[`%s`](#%s)", cf, cf)
	}

	module := cf[:i]
	symbol := cf[i+1:]
	if module == "builtin" {
		// The `builtin:` namespace is treated differently than other
		// namespaces with regard to how hash tag references are named.
		return fmt.Sprintf("[`%s`](%s.html/#%s)", cf, module, symbol)
	} else {
		return fmt.Sprintf("[`%s`](%s.html/#%s%s)", cf, module, module, symbol)
	}
}

func expandCf(line string) string {
	i := strings.Index(line, cf)
	if i < 0 {
		return line
	}
	targets := strings.Split(line[i+len(cf):], " ")
	var buf bytes.Buffer
	buf.WriteString("See also")
	for i, target := range targets {
		if i == 0 {
			buf.WriteString(" ")
		} else if i == len(targets)-1 {
			buf.WriteString(" and ")
		} else {
			buf.WriteString(", ")
		}
		buf.WriteString(makeLink(target))
	}
	buf.WriteString(".")
	return buf.String()
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
