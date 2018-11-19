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
			"Type":     List,
			"Elements": []ast{}}},
		ast{"Compound/Indexing/Primary", fs{
			"Type":     List,
			"Elements": []ast{}}},
		ast{"Compound/Indexing/Primary", fs{
			"Type":     List,
			"Elements": []string{"1"}}},
		ast{"Compound/Indexing/Primary", fs{
			"Type":     List,
			"Elements": []string{"2"}}},
		ast{"Compound/Indexing/Primary", fs{
			"Type":     List,
			"Elements": []string{"3"}}},
		ast{"Compound/Indexing/Primary", fs{
			"Type":     List,
			"Elements": []string{"4", "5", "6", "7"}}},
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
	{"a []{} [ ]{ } []{ echo 233 } [ x y ]{puts $x $y} { put haha}", a(
		ast{"Compound/Indexing/Primary", fs{
			"Type": Lambda, "Elements": []ast{}, "Chunk": "",
		}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": Lambda, "Elements": []ast{}, "Chunk": " ",
		}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": Lambda, "Elements": []ast{}, "Chunk": " echo 233 ",
		}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": Lambda, "Elements": []string{"x", "y"}, "Chunk": "puts $x $y",
		}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": Lambda, "Elements": []ast{}, "Chunk": " put haha",
		}},
	)},
	// Lambda with arguments and options
	{"a [a b &k=v]{}", a(
		ast{"Compound/Indexing/Primary", fs{
			"Type":     Lambda,
			"Elements": []string{"a", "b"},
			"MapPairs": []string{"&k=v"},
			"Chunk":    "",
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

	// Line continuation: "\\\n" is considered whitespace
	{"a b\\\nc", ast{
		"Chunk/Pipeline/Form", fs{"Head": "a", "Args": []string{"b", "c"}}}},
}

func TestParse(t *testing.T) {
	for _, tc := range goodCases {
		bn, err := AsChunk("[test]", tc.src)
		if err != nil {
			t.Errorf("Parse(%q) returns error: %v", tc.src, err)
		}
		err = checkParseTree(bn)
		if err != nil {
			t.Errorf("Parse(%q) returns bad parse tree: %v", tc.src, err)
			fmt.Fprintf(os.Stderr, "Parse tree of %q:\n", tc.src)
			PPrintParseTree(bn, os.Stderr)
		}
		err = checkAST(bn, tc.ast)
		if err != nil {
			t.Errorf("Parse(%q) returns bad AST: %v", tc.src, err)
			fmt.Fprintf(os.Stderr, "AST of %q:\n", tc.src)
			PPrintAST(bn, os.Stderr)
		}
	}
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
		_, err := AsChunk("[test]", tc.src)
		if err == nil {
			t.Errorf("Parse(%q) returns no error", tc.src)
			continue
		}
		posErr0 := err.(MultiError).Entries[0]
		if posErr0.Context.Begin != tc.pos {
			t.Errorf("Parse(%q) first error begins at %d, want %d. Errors are:%s\n", tc.src, posErr0.Context.Begin, tc.pos, err)
		}
	}
}
