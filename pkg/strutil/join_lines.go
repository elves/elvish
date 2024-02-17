package strutil

import "strings"

// JoinLines appends each line with a "\n" and joins all of them.
func JoinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
