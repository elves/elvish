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

func expandCf(line string) string {
	i := strings.Index(line, cf)
	if i < 0 {
		return line
	}
	targets := strings.Split(line[i+len(cf):], " ")
	var buf bytes.Buffer
	buf.WriteString("See also")
	for i, target := range targets {
		var sep string
		if i == 0 {
			sep = " "
		} else if i == len(targets)-1 {
			sep = " and "
		} else {
			sep = ", "
		}
		fmt.Fprintf(&buf, "%s[`%s`](#%s)", sep, target, target)
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
