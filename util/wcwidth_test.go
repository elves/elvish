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
