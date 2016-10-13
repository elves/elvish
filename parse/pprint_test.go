package parse

import (
	"bytes"
	"testing"
)

var pprintCases = []struct {
	src           string
	wantAST       string
	wantParseTree string
}{
	{"ls $x[0]$y[1];echo",
		`Chunk
  Pipeline/Form
    Compound/Indexing/Primary Type=Bareword Value="ls" IsRange=[]
    Compound
      Indexing
        Primary Type=Variable Value="x" IsRange=[]
        Array/Compound/Indexing/Primary Type=Bareword Value="0" IsRange=[]
      Indexing
        Primary Type=Variable Value="y" IsRange=[]
        Array/Compound/Indexing/Primary Type=Bareword Value="1" IsRange=[]
  Pipeline/Form/Compound/Indexing/Primary Type=Bareword Value="echo" IsRange=[]
`,
		`Chunk "ls $x[0]$y[1];echo" 0-18
  Pipeline/Form "ls $x[0]$y[1]" 0-13
    Compound/Indexing/Primary "ls" 0-2
    Sep " " 2-3
    Compound "$x[0]$y[1]" 3-13
      Indexing "$x[0]" 3-8
        Primary "$x" 3-5
        Sep "[" 5-6
        Array/Compound/Indexing/Primary "0" 6-7
        Sep "]" 7-8
      Indexing "$y[1]" 8-13
        Primary "$y" 8-10
        Sep "[" 10-11
        Array/Compound/Indexing/Primary "1" 11-12
        Sep "]" 12-13
  Sep ";" 13-14
  Pipeline/Form/Compound/Indexing/Primary "echo" 14-18
`},
}

func TestPprint(t *testing.T) {
	for _, tc := range pprintCases {
		n, err := Parse("[test]", tc.src)
		if err != nil {
			t.Error(err)
		}
		var b bytes.Buffer
		PprintAST(n, &b)
		ast := b.String()
		if b.String() != tc.wantAST {
			t.Errorf("PprintAST(%q):\n%s\nwant:\n%s", tc.src, ast, tc.wantAST)
		}
		b = bytes.Buffer{}
		PprintParseTree(n, &b)
		pt := b.String()
		if pt != tc.wantParseTree {
			t.Errorf("PprintParseTree(%q):\n%s\nwant:\n%s", tc.src, pt, tc.wantParseTree)
		}
	}
}
