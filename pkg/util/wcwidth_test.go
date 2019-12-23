package util

import (
	"testing"
)

var wcwidthTests = []struct {
	in     rune
	wanted int
}{
	{'\u0301', 0}, // Combining acute accent
	{'a', 1},
	{'Ω', 1},
	{'好', 2},
	{'か', 2},
}

func TestWcwidth(t *testing.T) {
	for _, tt := range wcwidthTests {
		out := Wcwidth(tt.in)
		if out != tt.wanted {
			t.Errorf("wcwidth(%q) => %v, want %v", tt.in, out, tt.wanted)
		}
	}
}

func TestOverrideWcwidth(t *testing.T) {
	r := '❱'
	oldw := Wcwidth(r)
	w := oldw + 1

	OverrideWcwidth(r, w)
	if Wcwidth(r) != w {
		t.Errorf("Wcwidth(%q) != %d after OverrideWcwidth", r, w)
	}
	UnoverrideWcwidth(r)
	if Wcwidth(r) != oldw {
		t.Errorf("Wcwidth(%q) != %d after UnoverrideWcwidth", r, oldw)
	}
}

func TestTrimWcwidth(t *testing.T) {
	if TrimWcwidth("abc", 2) != "ab" {
		t.Errorf("TrimWcwidth #1 fails")
	}
	if TrimWcwidth("你好", 3) != "你" {
		t.Errorf("TrimWcwidth #2 fails")
	}
}

func TestForceWcwidth(t *testing.T) {
	for i, c := range []struct {
		s    string
		w    int
		want string
	}{
		// Triming
		{"abc", 2, "ab"},
		{"你好", 2, "你"},
		// Padding
		{"abc", 4, "abc "},
		{"你好", 5, "你好 "},
		// Trimming and Padding
		{"你好", 3, "你 "},
	} {
		if got := ForceWcwidth(c.s, c.w); got != c.want {
			t.Errorf("ForceWcwidth #%d fails", i)
		}
	}
}

func TestTrimEachLineWcwidth(t *testing.T) {
	if TrimEachLineWcwidth("abcdefg\n你好", 3) != "abc\n你" {
		t.Errorf("TestTrimEachLineWcwidth fails")
	}
}
