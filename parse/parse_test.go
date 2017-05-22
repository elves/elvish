package parse

import (
	"fmt"
	"os"
	"testing"
)

func a(c ...interface{}) ast {
	// Shorthand used for checking Compound and levels beneath.
	return ast{"Chunk/Pipeline/Form", fs{"Head": "a", "Args": c}}
}

var goodCases = []struct {
	src string
	ast ast
}{
	// Chunk
	// Smoke test.
	{"a;b;c\n;d", ast{"Chunk", fs{"Pipelines": []string{"a", "b", "c", "d"}}}},
	// Empty chunk should have Pipelines=nil.
	{"", ast{"Chunk", fs{"Pipelines": nil}}},
	// Superfluous newlines and semicolons should not result in empty
	// pipelines.
	{"  ;\n\n  ls \t ;\n", ast{"Chunk", fs{"Pipelines": []string{"ls \t "}}}},

	// Pipeline
	{"a|b|c|d", ast{
		"Chunk/Pipeline", fs{"Forms": []string{"a", "b", "c", "d"}}}},
	// Newlines are allowed after pipes.
	{"a| \n \n b", ast{
		"Chunk/Pipeline", fs{"Forms": []string{"a", "b"}}}},
	// Comments.
	{"a#haha\nb#lala", ast{
		"Chunk", fs{"Pipelines": []string{"a", "b"}}}},

	// Form
	// Smoke test.
	{"ls x y", ast{"Chunk/Pipeline/Form", fs{
		"Head": "ls",
		"Args": []string{"x", "y"}}}},
	// Assignments.
	{"k=v k[a][b]=v {a,b[1]}=(ha)", ast{"Chunk/Pipeline/Form", fs{
		"Assignments": []string{"k=v", "k[a][b]=v", "{a,b[1]}=(ha)"}}}},
	// Temporary assignment.
	{"k=v k[a][b]=v a", ast{"Chunk/Pipeline/Form", fs{
		"Assignments": []string{"k=v", "k[a][b]=v"},
		"Head":        "a"}}},
	// Spacey assignment.
	{"k=v a b = c d", ast{"Chunk/Pipeline/Form", fs{
		"Assignments": []string{"k=v"},
		"Vars":        []string{"a", "b"},
		"Args":        []string{"c", "d"}}}},
	// Redirections
	{"a >b", ast{"Chunk/Pipeline/Form", fs{
		"Head": "a",
		"Redirs": []ast{
			{"Redir", fs{"Mode": Write, "Right": "b"}}},
	}}},
	// More redirections
	{"a >>b 2>b 3>&- 4>&1 5<c 6<>d", ast{"Chunk/Pipeline/Form", fs{
		"Head": "a",
		"Redirs": []ast{
			{"Redir", fs{"Mode": Append, "Right": "b"}},
			{"Redir", fs{"Left": "2", "Mode": Write, "Right": "b"}},
			{"Redir", fs{"Left": "3", "Mode": Write, "RightIsFd": true, "Right": "-"}},
			{"Redir", fs{"Left": "4", "Mode": Write, "RightIsFd": true, "Right": "1"}},
			{"Redir", fs{"Left": "5", "Mode": Read, "Right": "c"}},
			{"Redir", fs{"Left": "6", "Mode": ReadWrite, "Right": "d"}},
		},
	}}},
	// Exitus redirection
	{"a ?>$e", ast{"Chunk/Pipeline/Form", fs{
		"Head":        "a",
		"ExitusRedir": ast{"ExitusRedir", fs{"Dest": "$e"}},
	}}},
	// Options (structure of MapPair tested below with map)
	{"a &a=1 x &b=2", ast{"Chunk/Pipeline/Form", fs{
		"Head": "a",
		"Args": []string{"x"},
		"Opts": []string{"&a=1", "&b=2"},
	}}},

	// Compound
	{`a b"foo"?$c*'xyz'`, a(ast{"Compound", fs{
		"Indexings": []string{"b", `"foo"`, "?", "$c", "*", "'xyz'"}}})},

	// Indexing
	{"a $b[c][d][\ne\n]", a(ast{"Compound/Indexing", fs{
		"Head": "$b", "Indicies": []string{"c", "d", "\ne\n"},
	}})},

	// Primary
	//
	// Single quote
	{"a '''x''y'''", a(ast{"Compound/Indexing/Primary", fs{
		"Type": SingleQuoted, "Value": "'x'y'",
	}})},
	// Double quote
	{`a "b\^[\x1b\u548c\U0002CE23\123\n\t\\"`,
		a(ast{"Compound/Indexing/Primary", fs{
			"Type":  DoubleQuoted,
			"Value": "b\x1b\x1b\u548c\U0002CE23\123\n\t\\",
		}})},
	// Wildcard
	{"a * ?", a(
		ast{"Compound/Indexing/Primary", fs{"Type": Wildcard, "Value": "*"}},
		ast{"Compound/Indexing/Primary", fs{"Type": Wildcard, "Value": "?"}},
	)},
	// Variable
	{"a $x $&f", a(
		ast{"Compound/Indexing/Primary", fs{"Type": Variable, "Value": "x"}},
		ast{"Compound/Indexing/Primary", fs{"Type": Variable, "Value": "&f"}},
	)},
	// List
	{"a [] [ ] [1] [ 2] [3 ] [\n 4 \n5\n 6 7 \n]", a(
		ast{"Compound/Indexing/Primary", fs{
			"Type": List,
			"List": ""}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": List,
			"List": ""}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": List,
			"List": ast{"Array", fs{"Compounds": []string{"1"}}}}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": List,
			"List": ast{"Array", fs{"Compounds": []string{"2"}}}}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": List,
			"List": ast{"Array", fs{"Compounds": []string{"3"}}}}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": List,
			"List": ast{"Array", fs{
				"Compounds": []string{"4", "5", "6", "7"}}}}},
	)},
	// Semicolons in lists
	{"a [a b;c;d;]", a(
		ast{"Compound/Indexing/Primary", fs{
			"Type": List,
			"List": ast{"Array", fs{
				"Compounds":  []string{"a", "b", "c", "d"},
				"Semicolons": []int{2, 3, 4}}}}},
	)},
	// Map
	{"a [&k=v] [ &k=v] [&k=v ] [ &k=v ] [ &k= v] [&k= \n v] [\n&a=b &c=d \n &e=f\n\n]", a(
		ast{"Compound/Indexing/Primary", fs{
			"Type":     Map,
			"MapPairs": []ast{{"MapPair", fs{"Key": "k", "Value": "v"}}}}},
		ast{"Compound/Indexing/Primary", fs{
			"Type":     Map,
			"MapPairs": []ast{{"MapPair", fs{"Key": "k", "Value": "v"}}}}},
		ast{"Compound/Indexing/Primary", fs{
			"Type":     Map,
			"MapPairs": []ast{{"MapPair", fs{"Key": "k", "Value": "v"}}}}},
		ast{"Compound/Indexing/Primary", fs{
			"Type":     Map,
			"MapPairs": []ast{{"MapPair", fs{"Key": "k", "Value": "v"}}}}},
		ast{"Compound/Indexing/Primary", fs{
			"Type":     Map,
			"MapPairs": []ast{{"MapPair", fs{"Key": "k", "Value": "v"}}}}},
		ast{"Compound/Indexing/Primary", fs{
			"Type":     Map,
			"MapPairs": []ast{{"MapPair", fs{"Key": "k", "Value": "v"}}}}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": Map,
			"MapPairs": []ast{
				{"MapPair", fs{"Key": "a", "Value": "b"}},
				{"MapPair", fs{"Key": "c", "Value": "d"}},
				{"MapPair", fs{"Key": "e", "Value": "f"}},
			}}},
	)},
	// Empty map
	{"a [&] [ &] [& ] [ & ]", a(
		ast{"Compound/Indexing/Primary", fs{"Type": Map, "MapPairs": nil}},
		ast{"Compound/Indexing/Primary", fs{"Type": Map, "MapPairs": nil}},
		ast{"Compound/Indexing/Primary", fs{"Type": Map, "MapPairs": nil}},
		ast{"Compound/Indexing/Primary", fs{"Type": Map, "MapPairs": nil}},
	)},
	// Lambda
	{"a []{} [ ]{ } []{ echo 233 } [ $x $y ]{puts $x $y} { put $1}", a(
		ast{"Compound/Indexing/Primary", fs{
			"Type": Lambda, "List": "", "Chunk": "",
		}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": Lambda, "List": "", "Chunk": " ",
		}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": Lambda, "List": "", "Chunk": " echo 233 ",
		}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": Lambda, "List": "$x $y ", "Chunk": "puts $x $y",
		}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": Lambda, "List": nil, "Chunk": " put $1",
		}},
	)},
	// Output capture
	{"a () (b;c) (c\nd)", a(
		ast{"Compound/Indexing/Primary", fs{
			"Type": OutputCapture, "Chunk": ""}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": OutputCapture, "Chunk": ast{
				"Chunk", fs{"Pipelines": []string{"b", "c"}},
			}}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": OutputCapture, "Chunk": ast{
				"Chunk", fs{"Pipelines": []string{"c", "d"}},
			}}},
	)},
	// Output capture with backquotes
	{"a `` `b;c` `e>f`", a("``", "`b;c`", "`e>f`")},
	// Backquotes may be nested with unclosed parens and braces
	{"a `a (b `c`)` `d [`e`]`", a("`a (b `c`)`", "`d [`e`]`")},
	// Exitus capture
	{"a ?() ?(b;c)", a(
		ast{"Compound/Indexing/Primary", fs{
			"Type": ExceptionCapture, "Chunk": ""}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": ExceptionCapture, "Chunk": "b;c",
		}})},
	// Braced
	{"a {,a,c\ng\n}", a(
		ast{"Compound/Indexing/Primary", fs{
			"Type":   Braced,
			"Braced": []string{"", "a", "c", "g", ""}}})},
	// Tilde
	{"a ~xiaq/go", a(
		ast{"Compound", fs{
			"Indexings": []ast{
				{"Indexing/Primary", fs{"Type": Tilde, "Value": "~"}},
				{"Indexing/Primary", fs{"Type": Bareword, "Value": "xiaq/go"}},
			},
		}},
	)},
}

func TestParse(t *testing.T) {
	for _, tc := range goodCases {
		bn, err := Parse("[test]", tc.src)
		if err != nil {
			t.Errorf("Parse(%q) returns error: %v", tc.src, err)
		}
		err = checkParseTree(bn)
		if err != nil {
			t.Errorf("Parse(%q) returns bad parse tree: %v", tc.src, err)
			fmt.Fprintf(os.Stderr, "Parse tree of %q:\n", tc.src)
			PprintParseTree(bn, os.Stderr)
		}
		err = checkAST(bn, tc.ast)
		if err != nil {
			t.Errorf("Parse(%q) returns bad AST: %v", tc.src, err)
			fmt.Fprintf(os.Stderr, "AST of %q:\n", tc.src)
			PprintAST(bn, os.Stderr)
		}
	}
}

// checkParseTree checks whether the parse tree part of a Node is well-formed.
func checkParseTree(n Node) error {
	children := n.Children()
	if len(children) == 0 {
		return nil
	}

	// Parent pointers of all children should point to me.
	for i, ch := range children {
		if ch.Parent() != n {
			return fmt.Errorf("parent of child %d (%s) is wrong: %s", i, summary(ch), summary(n))
		}
	}

	// The Begin of the first child should be equal to mine.
	if children[0].Begin() != n.Begin() {
		return fmt.Errorf("gap between node and first child: %s", summary(n))
	}
	// The End of the last child should be equal to mine.
	nch := len(children)
	if children[nch-1].End() != n.End() {
		return fmt.Errorf("gap between node and last child: %s", summary(n))
	}
	// Consecutive children have consecutive position ranges.
	for i := 0; i < nch-1; i++ {
		if children[i].End() != children[i+1].Begin() {
			return fmt.Errorf("gap between child %d and %d of: %s", i, i+1, summary(n))
		}
	}

	// Check children recursively.
	for _, ch := range n.Children() {
		err := checkParseTree(ch)
		if err != nil {
			return err
		}
	}
	return nil
}

var badCases = []struct {
	src string
	pos int // expected Begin position of first error
}{
	// Empty form.
	{"a|", 2},
	// Unopened parens.
	{")", 0}, {"]", 0}, {"}", 0},
	// Unclosed parens.
	{"a (", 3}, {"a [", 3}, {"a {", 3},
	// Bogus ampersand.
	{"a & &", 4}, {"a [&", 4},
}

func TestParseError(t *testing.T) {
	for _, tc := range badCases {
		_, err := Parse("[test]", tc.src)
		if err == nil {
			t.Errorf("Parse(%q) returns no error", tc.src)
			continue
		}
		posErr0 := err.(*Error).Entries[0]
		if posErr0.Context.Begin != tc.pos {
			t.Errorf("Parse(%q) first error begins at %d, want %d. Errors are:%s\n", tc.src, posErr0.Context.Begin, tc.pos, err)
		}
	}
}
