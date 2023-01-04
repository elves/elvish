package strutil

// ChopLineEnding removes a line ending ("\r\n" or "\n") from the end of s. It
// returns s if it doesn't end with a line ending.
func ChopLineEnding(s string) string {
	if len(s) >= 2 && s[len(s)-2:] == "\r\n" { // Windows line ending
		return s[:len(s)-2]
	} else if len(s) >= 1 && s[len(s)-1] == '\n' { // Unix line ending
		return s[:len(s)-1]
	}
	return s
}

// ChopTerminator removes a specific terminator byte from the end of s. It
// returns s if it doesn't end with the specified terminator.
func ChopTerminator(s string, terminator byte) string {
	if len(s) >= 1 && s[len(s)-1] == terminator {
		return s[:len(s)-1]
	}
	return s
}
