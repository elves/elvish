package program

import "testing"

var quoteJSONTests = []struct {
	in   string
	want string
}{
	{`a`, `"a"`},
	{`"ab\c`, `"\"ab\\c"`},
	{"a\x19\x00", `"a\u0019\u0000"`},
}

func TestQuoteJSON(t *testing.T) {
	for _, test := range quoteJSONTests {
		out := quoteJSON(test.in)
		if out != test.want {
			t.Errorf("quoteJSON(%q) = %q, want %q", test.in, out, test.want)
		}
	}
}
