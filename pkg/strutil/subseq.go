package strutil

import "strings"

// HasSubseq determines whether s has t as its subsequence. A string t is a
// subsequence of a string s if and only if there is a possible sequence of
// steps of deleting characters from s that result in t.
func HasSubseq(s, t string) bool {
	for _, p := range t {
		i := strings.IndexRune(s, p)
		if i == -1 {
			return false
		}
		s = s[i+len(string(p)):]
	}
	return true
}
