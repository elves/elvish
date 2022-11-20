// Package elvdoc implements extraction of elvdoc, in-source documentation of
// Elvish variables and functions.
package elvdoc

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

var (
	// Groups:
	// 1. Name
	// 2. Signature (part inside ||)
	fnRegexp = regexp.MustCompile(`^fn +([^ ]+) +\{(?:\|([^|]*)\|)?`)
	// Groups:
	// 1. Name
	varRegexp = regexp.MustCompile(`^var +([^ ]+)`)
)

// Extract extracts elvdoc from Elvish source.
func Extract(r io.Reader) (fnDocs, varDocs map[string]string, err error) {
	fnDocs = make(map[string]string)
	varDocs = make(map[string]string)

	scanner := bufio.NewScanner(r)
	var commentLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# ") {
			commentLines = append(commentLines, line)
			continue
		}
		if m := fnRegexp.FindStringSubmatch(line); m != nil {
			name, sig := m[1], m[2]
			var sb strings.Builder
			writeUsage(&sb, name, sig)
			if len(commentLines) > 0 {
				sb.WriteByte('\n')
				writeCommentContent(&sb, commentLines)
			}
			fnDocs[name] = sb.String()
		} else if m := varRegexp.FindStringSubmatch(line); m != nil {
			name := m[1]
			var sb strings.Builder
			writeCommentContent(&sb, commentLines)
			varDocs[name] = sb.String()
		}
		commentLines = commentLines[:0]
	}

	return fnDocs, varDocs, scanner.Err()
}

func writeUsage(sb *strings.Builder, name, sig string) {
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
}

func writeCommentContent(sb *strings.Builder, lines []string) string {
	for _, line := range lines {
		// Every line starts with "# "
		sb.WriteString(line[2:])
		sb.WriteByte('\n')
	}
	return sb.String()
}

func Format(r io.Reader, w io.Writer) error {
	return nil
}
