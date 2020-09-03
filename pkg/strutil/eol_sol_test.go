package strutil

import "testing"

var EOLSOLTests = []struct {
	s                         string
	wantFirstEOL, wantLastSOL int
}{
	{"0", 1, 0},
	{"\n12", 0, 1},
	{"01\n", 2, 3},
	{"01\n34", 2, 3},
}

func TestEOLSOL(t *testing.T) {
	for _, tc := range EOLSOLTests {
		eol := FindFirstEOL(tc.s)
		if eol != tc.wantFirstEOL {
			t.Errorf("FindFirstEOL(%q) => %d, want %d", tc.s, eol, tc.wantFirstEOL)
		}
		sol := FindLastSOL(tc.s)
		if sol != tc.wantLastSOL {
			t.Errorf("FindLastSOL(%q) => %d, want %d", tc.s, sol, tc.wantLastSOL)
		}
	}
}
