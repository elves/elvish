package strutil

import "testing"

var hasSubseqTests = []struct {
	s, t string
	want bool
}{
	{"", "", true},
	{"a", "", true},
	{"a", "a", true},
	{"ab", "a", true},
	{"ab", "b", true},
	{"abc", "ac", true},
	{"abcdefg", "bg", true},
	{"abcdefg", "ga", false},
	{"foo lorem ipsum", "f l i", true},
	{"foo lorem ipsum", "oo o pm", true},
	{"你好世界", "好", true},
	{"你好世界", "好界", true},
}

func TestHasSubseq(t *testing.T) {
	for _, test := range hasSubseqTests {
		if b := HasSubseq(test.s, test.t); b != test.want {
			t.Errorf("HasSubseq(%q, %q) = %v, want %v", test.s, test.t, b, test.want)
		}
	}
}
