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
	{"k=v a b @rest = c d", ast{"Chunk/Pipeline/Form", fs{
		"Assignments": []string{"k=v"},
		"Vars":        []string{"a", "b", "@rest"},
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

	// Bareword.
	{"a foo", a(ast{"Compound/Indexing/Primary", fs{
		"Type": Bareword, "Value": "foo",
	}})},

	// Bareword, with all allowed symbols.
	{"a ./\\@%+!=,", a(ast{"Compound/Indexing/Primary", fs{
		"Type": Bareword, "Value": "./\\@%+!=,",
	}})},

	// Single quote
	{"a '''x''y'''", a(ast{"Compound/Indexing/Primary", fs{
		"Type": SingleQuoted, "Value": "'x'y'",
	}})},
	// Double quote
	{`a "[\c?\c@\cI\^I\^[]"`, // control char sequences
		a(ast{"Compound/Indexing/Primary", fs{
			"Type":  DoubleQuoted,
			"Value": "[\x7f\x00\t\t\x1b]",
		}})},

	{`a "[\n\t\a\v\\\"]"`, // single char sequences
		a(ast{"Compound/Indexing/Primary", fs{
			"Type":  DoubleQuoted,
			"Value": "[\n\t\a\v\\\"]",
		}})},

	{`a "b\^[\x1b\u548c\U0002CE23\123\n\t\\"`, // numeric sequences
		a(ast{"Compound/Indexing/Primary", fs{
			"Type":  DoubleQuoted,
			"Value": "b\x1b\x1b\u548c\U0002CE23\123\n\t\\",
		}})},
	// Wildcard
	{"a * ? ** ??", a(
		ast{"Compound/Indexing/Primary", fs{"Type": Wildcard, "Value": "*"}},
		ast{"Compound/Indexing/Primary", fs{"Type": Wildcard, "Value": "?"}},
		ast{"Compound/Indexing/Primary", fs{"Type": Wildcard, "Value": "**"}},
		ast{"Compound", fs{"Indexings": []string{"?", "?"}}},
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
	// Tilde and wildcard
	{"a ~xiaq/*.go", a(
		ast{"Compound", fs{
			"Indexings": []ast{
				{"Indexing/Primary", fs{"Type": Tilde, "Value": "~"}},
				{"Indexing/Primary", fs{"Type": Bareword, "Value": "xiaq/"}},
				{"Indexing/Primary", fs{"Type": Wildcard, "Value": "*"}},
				{"Indexing/Primary", fs{"Type": Bareword, "Value": ".go"}},
			},
		}},
	)},

	// Line continuation: "^\n" is considered whitespace
	{"a b^\nc", ast{
		"Chunk/Pipeline/Form", fs{"Head": "a", "Args": []string{"b", "c"}}}},

	// Carriage returns are normally treated the same as newlines:
	// Separating pipelines in a chunk
	{"a\rb", ast{"Chunk", fs{"Pipelines": []string{"a", "b"}}}},
	{"a\r\nb", ast{"Chunk", fs{"Pipelines": []string{"a", "b"}}}},
	// Whitespace padding in lambdas
	{"a { \rfoo\r\nbar }", a(
		ast{"Compound/Indexing/Primary",
			fs{"Type": Lambda, "Chunk": " \rfoo\r\nbar "}},
	)},
	// Separating elements in lists
	{"a [a\rb]", a(
		ast{"Compound/Indexing/Primary", fs{
			"Type":     List,
			"Elements": []string{"a", "b"}}})},

	// However, in line continuations, \r\n is treated as a single newline
	{"a b^\r\nc", ast{
		"Chunk/Pipeline/Form", fs{"Head": "a", "Args": []string{"b", "c"}}}},
	// But a lone \r also works
	{"a b^\rc", ast{
		"Chunk/Pipeline/Form", fs{"Head": "a", "Args": []string{"b", "c"}}}},

	// Comments in chunks.
	{"a#haha\nb#lala", ast{
		"Chunk", fs{"Pipelines": []ast{
			{"Pipeline/Form", fs{"Head": "a"}},
			{"Pipeline/Form", fs{"Head": "b"}},
		}}}},
	// Comments in lists.
	{"a [a#haha\nb]", a(
		ast{"Compound/Indexing/Primary", fs{
			"Type":     List,
			"Elements": []string{"a", "b"},
		}},
	)},
}

func TestParse(t *testing.T) {
	for _, tc := range goodCases {
		src := SourceForTest(tc.src)
		tree, err := Parse(src)
		if err != nil {
			t.Errorf("Parse(%q) returns error: %v", tc.src, err)
		}
		if tree.Source != src {
			t.Errorf("Parse(%q) returns source %v, want %v", tc.src, tree.Source, src)
		}
		err = checkParseTree(tree.Root)
		if err != nil {
			t.Errorf("Parse(%q) returns bad parse tree: %v", tc.src, err)
			fmt.Fprintf(os.Stderr, "Parse tree of %q:\n", tc.src)
			pprintParseTree(tree.Root, os.Stderr)
		}
		err = checkAST(tree.Root, tc.ast)
		if err != nil {
			t.Errorf("Parse(%q) returns bad AST: %v", tc.src, err)
			fmt.Fprintf(os.Stderr, "AST of %q:\n", tc.src)
			pprintAST(tree.Root, os.Stderr)
		}
	}
}

var parseErrorTests = []struct {
	src      string
	errPart  string
	errAtEnd bool
	errMsg   string
}{
	// Empty form.
	{src: "a|", errAtEnd: true, errMsg: "should be form"},
	// Unopened parens.
	{src: ")", errPart: ")", errMsg: "unexpected rune ')'"},
	{src: "]", errPart: "]", errMsg: "unexpected rune ']'"},
	{src: "}", errPart: "}", errMsg: "unexpected rune '}'"},
	// Unclosed parens.
	{src: "a (", errAtEnd: true, errMsg: "should be ')'"},
	{src: "a [", errAtEnd: true, errMsg: "should be ']'"},
	{src: "a {", errAtEnd: true, errMsg: "should be ',' or '}'"},
	// Bogus ampersand in form.
	{src: "a & &", errPart: "&", errMsg: "unexpected rune '&'"},
	// Bad assignment LHS.
	{src: "a'b' = x", errPart: "a'b'", errMsg: "bad assignment LHS"},
	{src: "$a = x", errPart: "$a", errMsg: "bad assignment LHS"},
	{src: "'' = x", errPart: "''", errMsg: "bad assignment LHS"},
	{src: "'<' = x", errPart: "'<'", errMsg: "bad assignment LHS"},
	// Chained assignment.
	{src: "a = b = c", errPart: "=", errMsg: "chained assignment not yet supported"},
	// No redirection source.
	{src: "a >", errAtEnd: true, errMsg: "should be a composite term representing filename"},
	{src: "a >&", errAtEnd: true, errMsg: "should be a composite term representing fd"},
	// Unmatched paren in indexing.
	{src: "a $a[0}", errPart: "}", errMsg: "should be ']'"},
	// Unterminated string.
	{src: "'a", errAtEnd: true, errMsg: "string not terminated"},
	{src: `"a`, errAtEnd: true, errMsg: "string not terminated"},
	// Bad escape sequence.
	{src: `a "\^` + "\t", errPart: "\t",
		errMsg: "invalid control sequence, should be a codepoint between 0x3F and 0x5F"},
	{src: `a "\xQQ"`, errPart: "Q", errMsg: "invalid escape sequence, should be hex digit"},
	{src: `a "\1ab"`, errPart: "a", errMsg: "invalid escape sequence, should be octal digit"},
	{src: `a "\i"`, errPart: "i", errMsg: "invalid escape sequence"},
	// Unterminated variable name.
	{src: "$", errAtEnd: true, errMsg: "should be variable name"},
	// Unmatched (.
	{src: "a (", errAtEnd: true, errMsg: "should be ')'"},
	// List-map hybrid.
	// TODO(xiaq): Add correct position information.
	{src: "a [a &k=v]", errAtEnd: true, errMsg: "cannot contain both list elements and map pairs"},
	// Unmatched {.
	{src: "{ a", errAtEnd: true, errMsg: "should be '}'"},
	// Unfinished line continuation.
	{src: `a ^`, errAtEnd: true, errMsg: "should be newline"},
}

func TestParseError(t *testing.T) {
	for _, test := range parseErrorTests {
		t.Run(test.src, func(t *testing.T) {
			_, err := Parse(SourceForTest(test.src))
			if err == nil {
				t.Fatalf("no error")
			}
			parseError := err.(*MultiError).Entries[0]
			r := parseError.Context

			if errPart := test.src[r.From:r.To]; errPart != test.errPart {
				t.Errorf("err part is %q, want %q", errPart, test.errPart)
			}
			if errAtEnd := r.From == len(test.src); errAtEnd != test.errAtEnd {
				t.Errorf("err at end is %v, want %v", errAtEnd, test.errAtEnd)
			}
			if errMsg := parseError.Message; errMsg != test.errMsg {
				t.Errorf("err message is %q, want %q", errMsg, test.errMsg)
			}
		})
	}
}
