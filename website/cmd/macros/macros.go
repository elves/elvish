// The macros program implements an ad-hoc preprocessor for Markdown files, used
// in Elvish's website.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var (
	repoPath   = flag.String("repo", "", "root of repo")
	elvdocPath = flag.String("elvdoc", "", "Path to the elvdoc binary")
)

func main() {
	flag.Parse()
	filter(os.Stdin, os.Stdout)
}

func filter(in io.Reader, out io.Writer) {
	f := filterer{}
	f.filter(in, out)
}

type filterer struct {
	module, path string
}

var macros = map[string]func(*filterer, string) string{
	"@module ":  (*filterer).expandModule,
	"@ttyshot ": (*filterer).expandTtyshot,
	"@cf ":      (*filterer).expandCf,
	"@dl ":      (*filterer).expandDl,
}

func (f *filterer) filter(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()
		for leader, expand := range macros {
			i := strings.Index(line, leader)
			if i >= 0 {
				line = line[:i] + expand(f, line[i+len(leader):])
				break
			}
		}
		fmt.Fprintln(out, line)
	}
	if f.module != "" {
		callElvdoc(out, f.module, f.path)
	}
}

func (f *filterer) expandModule(rest string) string {
	if *repoPath == "" || *elvdocPath == "" {
		log.Println("-repo and -elvdoc are required to expand @module ", rest)
		return ""
	}
	fields := strings.Fields(rest)
	switch len(fields) {
	case 1:
		f.module = fields[0]
		f.path = "pkg/mods/" + strings.ReplaceAll(f.module, "-", "")
	case 2:
		f.module = fields[0]
		f.path = fields[1]
	default:
		log.Println("bad macro: @module ", rest)
	}
	// Module doc will be added at end of file
	return fmt.Sprintf(
		"<a name='//apple_ref/cpp/Module/%s' class='dashAnchor'></a>", f.module)
}

func callElvdoc(out io.Writer, module, path string) {
	fullPath := filepath.Join(*repoPath, path)
	ns := module + ":"
	if module == "builtin" {
		ns = ""
	}

	cmd := exec.Command(*elvdocPath, "-ns", ns, fullPath)
	r, w := io.Pipe()
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	filterDone := make(chan struct{})
	go func() {
		filter(r, out)
		close(filterDone)
	}()
	err := cmd.Run()
	w.Close()
	<-filterDone
	r.Close()
	if err != nil {
		log.Fatalln(err)
	}
}

func (f *filterer) expandTtyshot(name string) string {
	content, err := os.ReadFile(path.Join("ttyshot", name+".html"))
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf(`<pre class="ttyshot"><code>%s</code></pre>`,
		bytes.Replace(content, []byte("\n"), []byte("<br>"), -1))
}

func (f *filterer) expandCf(rest string) string {
	targets := strings.Split(rest, " ")
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
	// A link to a non-builtin page.
	return module + ".html#" + target
}

func (f *filterer) expandDl(rest string) string {
	fields := strings.SplitN(rest, " ", 2)
	name := fields[0]
	url := name
	if len(fields) == 2 {
		url = fields[1]
	}
	return fmt.Sprintf(`<a href="https://dl.elv.sh/%s">%s</a>`, url, name)
}
