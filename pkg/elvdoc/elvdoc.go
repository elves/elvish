// Package elvdoc extracts doc comments of Elvish variables and functions.
package elvdoc

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

// Docs records doc comments.
type Docs struct {
	Fns  []Entry
	Vars []Entry
}

// Entry is a doc comment entry.
type Entry struct {
	Name string
	// ID to use in HTML
	ID      string
	Content string
}

var (
	// Groups:
	// 1. Name
	// 2. Signature (part inside ||)
	fnRegexp = regexp.MustCompile(`^fn +([^ ]+) +\{(?:\|([^|]*)\|)?`)
	// Groups:
	// 1. Name
	varRegexp = regexp.MustCompile(`^var +([^ ]+)`)
	// Groups:
	// 1. Names
	fnNoSigRegexp = regexp.MustCompile(`^#doc:fn +(.+)`)
	// Groups:
	// 1. Name
	idRegexp = regexp.MustCompile(`^#doc:id +(.+)`)
)

const showUnstable = "#doc:show-unstable"

// Extract extracts elvdoc from Elvish source.
func Extract(r io.Reader, prefix string) (Docs, error) {
	var docs Docs
	var block docBlock
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "#" {
			block.lines = append(block.lines, "")
		} else if strings.HasPrefix(line, "# ") {
			block.lines = append(block.lines, line[2:])
		} else if line == showUnstable {
			block.showUnstable = true
		} else if m := fnRegexp.FindStringSubmatch(line); m != nil {
			name, sig := m[1], m[2]
			content := fnUsage(prefix+name, sig)
			id := ""
			showUnstable := false
			if len(block.lines) > 0 {
				var blockContent string
				id, blockContent, showUnstable = block.consume()
				content += "\n" + blockContent
			}
			if showUnstable || !unstable(name) {
				docs.Fns = append(docs.Fns, Entry{name, id, content})
			}
		} else if m := varRegexp.FindStringSubmatch(line); m != nil {
			name := m[1]
			id, content, showUnstable := block.consume()
			if showUnstable || !unstable(name) {
				docs.Vars = append(docs.Vars, Entry{name, id, content})
			}
		} else if m := fnNoSigRegexp.FindStringSubmatch(line); m != nil {
			name := m[1]
			id, content, _ := block.consume()
			docs.Fns = append(docs.Fns, Entry{name, id, content})
		} else if m := idRegexp.FindStringSubmatch(line); m != nil {
			block.id = m[1]
		} else {
			block.consume()
		}
	}

	return docs, scanner.Err()
}

func fnUsage(name, sig string) string {
	var sb strings.Builder
	sb.WriteString("```elvish\n")
	sb.WriteString(name)
	for _, field := range strings.Fields(sig) {
		sb.WriteByte(' ')
		if strings.HasPrefix(field, "&") {
			sb.WriteString(field)
		} else if strings.HasPrefix(field, "@") {
			sb.WriteString("$" + field[1:] + "...")
		} else {
			sb.WriteString("$" + field)
		}
	}
	sb.WriteString("\n```\n")
	return sb.String()
}

type docBlock struct {
	id           string
	lines        []string
	showUnstable bool
}

func (db *docBlock) consume() (id, content string, showUnstable bool) {
	id = db.id
	db.id = ""

	var sb strings.Builder
	for _, line := range db.lines {
		sb.WriteString(line)
		sb.WriteByte('\n')
	}
	db.lines = db.lines[:0]

	showUnstable = db.showUnstable
	db.showUnstable = false

	return id, sb.String(), showUnstable
}

func unstable(s string) bool { return s != "-" && strings.HasPrefix(s, "-") }
