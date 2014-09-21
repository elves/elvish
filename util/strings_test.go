package util

import "testing"

var findContextTests = []struct {
	text          string
	pos           int
	lineno, colno int
	line          string
}{
	{"a\nb", 2, 1, 0, "b"},
}

func TestFindContext(t *testing.T) {
	for _, tt := range findContextTests {
		lineno, colno, line := FindContext(tt.text, tt.pos)
		if lineno != tt.lineno || colno != tt.colno || line != tt.line {
			t.Errorf("FindContext(%v, %v) => (%v, %v, %v), want (%v, %v, %v)",
				lineno, colno, line, tt.lineno, tt.colno, tt.line)
		}
	}
}

var SubstringByRuneTests = []struct {
	s         string
	low, high int
	wantedStr string
	wantedErr error
}{
	{"Hello world", 1, 4, "ell", nil},
	{"你好世界", 0, 0, "", nil},
	{"你好世界", 1, 1, "", nil},
	{"你好世界", 1, 2, "好", nil},
	{"你好世界", 1, 4, "好世界", nil},
	{"你好世界", -1, -1, "", IndexOutOfRange},
	{"你好世界", 0, 5, "", IndexOutOfRange},
	{"你好世界", 5, 5, "", IndexOutOfRange},
}

func TestSubstringByRune(t *testing.T) {
	for _, tt := range SubstringByRuneTests {
		s, e := SubstringByRune(tt.s, tt.low, tt.high)
		if s != tt.wantedStr || e != tt.wantedErr {
			t.Errorf("SubstringByRune(%q, %v, %d) => (%q, %v), want (%q, %v)",
				tt.s, tt.low, tt.high, s, e, tt.wantedStr, tt.wantedErr)
		}
	}
}

var NthRuneTests = []struct {
	s          string
	n          int
	wantedRune rune
	wantedErr  error
}{
	{"你好世界", -1, 0, IndexOutOfRange},
	{"你好世界", 0, '你', nil},
	{"你好世界", 4, 0, IndexOutOfRange},
}

func TestNthRune(t *testing.T) {
	for _, tt := range NthRuneTests {
		r, e := NthRune(tt.s, tt.n)
		if r != tt.wantedRune || e != tt.wantedErr {
			t.Errorf("NthRune(%q, %v) => (%q, %v), want (%q, %v)",
				tt.s, tt.n, r, e, tt.wantedRune, tt.wantedErr)
		}
	}
}
