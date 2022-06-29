// The highlight program highlights Elvish code fences in Markdown.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"html"
	"log"
	"os"
	"strings"

	"src.elv.sh/pkg/edit/highlight"
)

var (
	scanner     = bufio.NewScanner(os.Stdin)
	lineno      = 0
	highlighter = highlight.NewHighlighter(highlight.Config{})
	raw         = false
)

func scan() bool {
	lineno++
	return scanner.Scan()
}

func main() {
	flags := flag.NewFlagSet("", flag.ExitOnError)
	var rawFlag = flags.Bool("raw", false, "raw output (for generating help text)")
	err := flags.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	args := flags.Args()
	if len(args) > 0 {
		log.Fatal("expected zero non-flag arguments")
	}
	raw = *rawFlag

	for scan() {
		line := scanner.Text()
		trimmed := strings.TrimLeft(line, " ")
		indent := line[:len(line)-len(trimmed)]
		if trimmed == "```elvish" || trimmed == "```elvish-usage" || trimmed == "```elvish-bad" {
			bad := trimmed == "```elvish-bad"
			highlighted := convert(collectFenced(indent), bad)
			if raw {
				usage := trimmed == "```elvish-usage"
				outputRawCodeBlock(highlighted, usage)
			} else {
				fmt.Printf("%s<pre><code>%s</code></pre>\n", indent, highlighted)
			}
		} else if trimmed == "```elvish-transcript" {
			highlighted := convertTranscript(collectFenced(indent))
			if raw {
				outputRawCodeBlock(highlighted, false)
			} else {
				fmt.Printf("%s<pre><code>%s</code></pre>\n", indent, highlighted)
			}
		} else {
			fmt.Println(line)
		}
	}
}

// outputRawCodeBlock breaks a highlighted block of code on newline boundaries and outputs each
// line. If the code block is "usage" text it is not indented; otherwise, it has a four space indent
// to make reading the output of the Elvish `help` command easier.
func outputRawCodeBlock(code string, usage bool) {
	if usage {
		fmt.Println("Usage:")
		fmt.Println("")
	}
	for _, line := range strings.Split(code, "\n") {
		fmt.Println("   ", line) // this indentation inhibits reflow by the `glow` renderer
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
			if raw {
				buf.WriteString(ps1)
				buf.WriteString(convert(elvishBuf.String(), false))
			} else {
				buf.WriteString(html.EscapeString(ps1))
				buf.WriteString(convert(elvishBuf.String(), false))
			}
		} else {
			if raw {
				buf.WriteString(line)
				buf.WriteString("\n")
			} else {
				buf.WriteString(html.EscapeString(line))
				buf.WriteString("<br>")
			}
		}
	}
	return buf.String()
}

func convert(text string, bad bool) string {
	highlighted, errs := highlighter.Get(text)
	if len(errs) != 0 && !bad {
		log.Printf("parsing %q: %v", text, errs)
	}

	if raw {
		return highlighted.String()
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
