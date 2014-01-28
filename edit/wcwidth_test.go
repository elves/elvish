package edit

import (
	"testing"
)

var wcwidthTests = []struct{
	in rune
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
		out := wcwidth(tt.in)
		if out != tt.wanted {
			t.Errorf("wcwidth(%q) => %v, want %v", tt.in, out, tt.wanted)
		}
	}
}
