package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html"
	"log"
	"os"
	"strings"

	"github.com/elves/elvish/edit/highlight"
)

var (
	scanner     = bufio.NewScanner(os.Stdin)
	lineno      = 0
	highlighter = highlight.NewHighlighter(highlight.Config{})
)

func scan() bool {
	lineno++
	return scanner.Scan()
}

func main() {
	for scan() {
		line := scanner.Text()
		trimmed := strings.TrimLeft(line, " ")
		indent := line[:len(line)-len(trimmed)]
		if trimmed == "```elvish" || trimmed == "```elvish-bad" {
			bad := trimmed == "```elvish-bad"
			highlighted := convert(collectFenced(indent), bad)
			fmt.Printf("%s<pre><code>%s</code></pre>\n", indent, highlighted)
		} else if trimmed == "```elvish-transcript" {
			highlighted := convertTranscript(collectFenced(indent))
			fmt.Printf("%s<pre><code>%s</code></pre>\n", indent, highlighted)
		} else {
			fmt.Println(line)
		}
	}
}

func collectFenced(indent string) string {
	var buf bytes.Buffer
	for scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, indent) {
			log.Fatalf("bad indent of line %d: %q", lineno, line)
		}
		unindented := line[len(indent):]
		if unindented == "```" {
			break
		}
		if buf.Len() > 0 {
			buf.WriteRune('\n')
		}
		buf.WriteString(unindented)
	}
	return buf.String()
}

const (
	ps1 = "~> "
	ps2 = "   "
)

func convertTranscript(transcript string) string {
	scanner := bufio.NewScanner(bytes.NewBufferString(transcript))
	var buf bytes.Buffer

	overread := false
	var line string
	for overread || scanner.Scan() {
		if overread {
			overread = false
		} else {
			line = scanner.Text()
		}
		if strings.HasPrefix(line, ps1) {
			elvishBuf := bytes.NewBufferString(line[len(ps1):] + "\n")
			for scanner.Scan() {
				line = scanner.Text()
				if strings.HasPrefix(line, ps2) {
					elvishBuf.WriteString(line + "\n")
				} else {
					overread = true
					break
				}
			}
			buf.WriteString(html.EscapeString(ps1))
			buf.WriteString(convert(elvishBuf.String(), false))
		} else {
			buf.WriteString(html.EscapeString(line))
			buf.WriteString("<br>")
		}
	}
	return buf.String()
}

func convert(text string, bad bool) string {
	highlighted, errs := highlighter.Get(text)
	if len(errs) != 0 && !bad {
		log.Printf("parsing %q: %v", text, errs)
	}

	var buf strings.Builder

	for _, seg := range highlighted {
		var classes []string
		for _, sgrCode := range strings.Split(seg.Style.SGR(), ";") {
			classes = append(classes, "sgr-"+sgrCode)
		}
		jointClass := strings.Join(classes, " ")
		if len(jointClass) > 0 {
			fmt.Fprintf(&buf, `<span class="%s">`, jointClass)
		}
		for _, r := range seg.Text {
			if r == '\n' {
				buf.WriteString("<br>")
			} else {
				buf.WriteString(html.EscapeString(string(r)))
			}
		}
		if len(jointClass) > 0 {
			buf.WriteString("</span>")
		}
	}

	return buf.String()
}
