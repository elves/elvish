package util

import "unicode/utf8"

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
