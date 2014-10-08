package parse

import "testing"

var lexTests = []struct {
	in  string
	out []Item
}{
	// Literals
	{"a `b``c` \"d\\\"e\"", []Item{
		{ItemBare, 0, "a", ItemAmbiguous},
		{ItemSpace, 1, " ", ItemAmbiguous},
		{ItemSingleQuoted, 2, "`b``c`", ItemAmbiguous},
		{ItemSpace, 8, " ", ItemAmbiguous},
		{ItemDoubleQuoted, 9, `"d\"e"`, ItemTerminated},
	}},
	// Comment
	{"a #b\nc", []Item{
		{ItemBare, 0, "a", ItemAmbiguous},
		{ItemSpace, 1, " ", ItemAmbiguous},
		{ItemSpace, 2, "#b", ItemAmbiguous},
		{ItemEndOfLine, 4, "\n", ItemTerminated},
		{ItemBare, 5, "c", ItemAmbiguous},
	}},
}

func TestLex(t *testing.T) {
tt:
	for _, tt := range lexTests {
		l := Lex("<test case>", tt.in)
		var out []Item
		for {
			token := l.NextItem()
			if token.Typ == ItemEOF {
				break
			}
			out = append(out, token)
		}
		for i := range out {
			if out[i] != tt.out[i] {
				t.Errorf("%q => %#v, want %#v", tt.in, out, tt.out)
				continue tt
			}
		}
	}
}
