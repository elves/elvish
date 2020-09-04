package strutil

// ChopLineEnding removes a line ending ("\r\n" or "\n") from the end of s. It
// returns itself if it doesn't end with a line ending.
func ChopLineEnding(s string) string {
	if len(s) >= 2 && s[len(s)-2:] == "\r\n" {
		return s[:len(s)-2]
	} else if len(s) >= 1 && s[len(s)-1] == '\n' {
		return s[:len(s)-1]
	}
	return s
}
