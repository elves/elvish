package util

import "unicode/utf8"

// HasSubseq determines whether s has t as its subsequence. A string t is a
// subsequence of a string s if and only if there is a possible sequence of
// steps of deleting characters from s that result in t.
func HasSubseq(s, t string) bool {
	i, j := 0, 0
	for i < len(s) && j < len(t) {
		s0, di := utf8.DecodeRuneInString(s[i:])
		t0, dj := utf8.DecodeRuneInString(t[j:])
		i += di
		if s0 == t0 {
			j += dj
		}
	}
	return j == len(t)
}
