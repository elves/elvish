package util

import "testing"

var findContextTests = []struct{
	text string
	pos int
	lineno, colno int
	line string
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
