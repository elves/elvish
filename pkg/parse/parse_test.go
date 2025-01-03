package parse

import (
	"fmt"
	"os"
	"testing"
)

func a(c ...any) ast {
	// Shorthand used for checking Compound and levels beneath.
	return ast{"Chunk/Pipeline/Form", fs{"Head": "a", "Args": c}}
}

var testCases = []struct {
	name string
	code string
	node Node
	want ast

	wantErrPart  string
	wantErrAtEnd bool
	wantErrMsg   string
}{
	// Chunk
	{
		name: "empty chunk",
		code: "",
		node: &Chunk{},
		want: ast{"Chunk", fs{"Pipelines": nil}},
	},
	{
		name: "multiple pipelines separated by newlines and semicolons",
		code: "a;b;c\n;d",
		node: &Chunk{},
		want: ast{"Chunk", fs{"Pipelines": []string{"a", "b", "c", "d"}}},
	},
	{
		name: "extra newlines and semicolons do not result in empty pipelines",
		code: "  ;\n\n  ls \t ;\n",
		node: &Chunk{},
		want: ast{"Chunk", fs{"Pipelines": []string{"ls \t "}}},
	},

	// Pipeline
	{
		name: "pipeline",
		code: "a|b|c|d",
		node: &Pipeline{},
		want: ast{"Pipeline", fs{"Forms": []string{"a", "b", "c", "d"}}},
	},
	{
		name: "newlines after pipes are allowed",
		code: "a| \n \n b",
		node: &Pipeline{},
		want: ast{"Pipeline", fs{"Forms": []string{"a", "b"}}},
	},

	{
		name:         "no form after pipe",
		code:         "a|",
		node:         &Chunk{},
		wantErrAtEnd: true,
		wantErrMsg:   "should be form",
	},

	// Form
	{
		name: "command form",
		code: "ls x y",
		node: &Form{},
		want: ast{"Form", fs{
			"Head": "ls",
			"Args": []string{"x", "y"}}},
	},
	{
		name: "redirection",
		code: "a >b",
		node: &Form{},
		want: ast{"Form", fs{
			"Head": "a",
			"Redirs": []ast{
				{"Redir", fs{"Mode": Write, "Right": "b"}}},
		}},
	},
	{
		name: "advanced redirections",
		code: "a >>b 2>b 3>&- 4>&1 5<c 6<>d",
		node: &Form{},
		want: ast{"Form", fs{
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
	{
		name: "command options",
		code: "a &a=1 x &b=2",
		node: &Form{},
		want: ast{"Form", fs{
			"Head": "a",
			"Args": []string{"x"},
			"Opts": []string{"&a=1", "&b=2"},
		}},
		// More tests for MapPair below with map syntax
	},

	{
		name:        "bogus ampersand in command form",
		code:        "a & &",
		node:        &Chunk{},
		wantErrPart: "&",
		wantErrMsg:  "unexpected rune '&'",
	},
	{
		name:         "no filename redirection source",
		code:         "a >",
		node:         &Chunk{},
		wantErrAtEnd: true,
		wantErrMsg:   "should be a composite term representing filename",
	},
	{
		name:         "no FD direction source",
		code:         "a >&",
		node:         &Chunk{},
		wantErrAtEnd: true,
		wantErrMsg:   "should be a composite term representing fd",
	},

	// Filter
	{
		name: "empty filter",
		code: "",
		node: &Filter{},
		want: ast{"Filter", fs{}},
	},
	{
		name: "filter with arguments",
		code: "foo bar",
		node: &Filter{},
		want: ast{"Filter", fs{"Args": []string{"foo", "bar"}}},
	},
	{
		name: "filter with options",
		code: "&foo=bar &lorem=ipsum",
		node: &Filter{},
		want: ast{"Filter", fs{"Opts": []string{"&foo=bar", "&lorem=ipsum"}}},
	},
	{
		name: "filter mixing arguments and options",
		code: "foo &a=b bar &x=y",
		node: &Filter{},
		want: ast{"Filter", fs{
			"Args": []string{"foo", "bar"},
			"Opts": []string{"&a=b", "&x=y"}}},
	},
	{
		name: "filter with leading and trailing whitespaces",
		code: "  foo  ",
		node: &Filter{},
		want: ast{"Filter", fs{"Args": []string{"foo"}}},
	},

	// Compound
	{
		name: "compound expression",
		code: `b"foo"?$c*'xyz'`,
		node: &Compound{},
		want: ast{"Compound", fs{
			"Indexings": []string{"b", `"foo"`, "?", "$c", "*", "'xyz'"}}},
	},

	// Indexing
	{
		name: "indexing expression",
		code: "$b[c][d][\ne\n]",
		node: &Indexing{},
		want: ast{"Indexing", fs{
			"Head": "$b", "Indices": []string{"c", "d", "\ne\n"}}},
	},
	{
		name: "indexing expression with empty index",
		code: "$a[]",
		node: &Indexing{},
		want: ast{"Indexing", fs{
			"Head": "$a", "Indices": []string{""}}},
	},

	// Primary
	{
		name: "bareword",
		code: "foo",
		node: &Primary{},
		want: ast{"Primary", fs{"Type": Bareword, "Value": "foo"}},
	},
	{
		name: "bareword with all allowed symbols",
		code: "./\\@%+!=,",
		node: &Primary{},
		want: ast{"Primary", fs{"Type": Bareword, "Value": "./\\@%+!=,"}},
	},
	{
		name: "single-quoted string",
		code: "'''x''y'''",
		node: &Primary{},
		want: ast{"Primary", fs{"Type": SingleQuoted, "Value": "'x'y'"}},
	},
	{
		name: "double-quoted string with control char escape sequences",
		code: `"[\c?\c@\cI\^I\^[]"`,
		node: &Primary{},
		want: ast{"Primary", fs{
			"Type":  DoubleQuoted,
			"Value": "[\x7f\x00\t\t\x1b]",
		}},
	},
	{
		name: "double-quoted string with single-char escape sequences",
		code: `"[\n\t\a\v\\\"]"`,
		node: &Primary{},
		want: ast{"Primary", fs{
			"Type":  DoubleQuoted,
			"Value": "[\n\t\a\v\\\"]",
		}},
	},
	{
		name: "double-quoted string with numerical escape sequences for codepoints",
		code: `"b\^[\u548c\U0002CE23\n\t\\"`,
		node: &Primary{},
		want: ast{"Primary", fs{
			"Type":  DoubleQuoted,
			"Value": "b\x1b\u548c\U0002CE23\n\t\\",
		}},
	},
	{
		name: "double-quoted string with numerical escape sequences for bytes",
		code: `"\123\321 \x7f\xff"`,
		node: &Primary{},
		want: ast{"Primary", fs{
			"Type":  DoubleQuoted,
			"Value": "\123\321 \x7f\xff",
		}},
	},
	{
		name: "wildcard",
		code: "a * ? ** ??",
		node: &Chunk{},
		want: a(
			ast{"Compound/Indexing/Primary", fs{"Type": Wildcard, "Value": "*"}},
			ast{"Compound/Indexing/Primary", fs{"Type": Wildcard, "Value": "?"}},
			ast{"Compound/Indexing/Primary", fs{"Type": Wildcard, "Value": "**"}},
			ast{"Compound", fs{"Indexings": []string{"?", "?"}}},
		),
	},
	{
		name: "variable",
		code: `a $x $'!@#' $"\n"`,
		node: &Chunk{},
		want: a(
			ast{"Compound/Indexing/Primary", fs{"Type": Variable, "Value": "x"}},
			ast{"Compound/Indexing/Primary", fs{"Type": Variable, "Value": "!@#"}},
			ast{"Compound/Indexing/Primary", fs{"Type": Variable, "Value": "\n"}},
		),
	},
	{
		name: "list",
		code: "a [] [ ] [1] [ 2] [3 ] [\n 4 \n5\n 6 7 \n]",
		node: &Chunk{},
		want: a(
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
		),
	},
	{
		name: "map",
		code: "a [&k=v] [ &k=v] [&k=v ] [ &k=v ] [ &k= v] [&k= \n v] [\n&a=b &c=d \n &e=f\n\n]",
		node: &Chunk{},
		want: a(
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
		),
	},
	{
		name: "empty map",
		code: "a [&] [ &] [& ] [ & ]",
		node: &Chunk{},
		want: a(
			ast{"Compound/Indexing/Primary", fs{"Type": Map, "MapPairs": nil}},
			ast{"Compound/Indexing/Primary", fs{"Type": Map, "MapPairs": nil}},
			ast{"Compound/Indexing/Primary", fs{"Type": Map, "MapPairs": nil}},
			ast{"Compound/Indexing/Primary", fs{"Type": Map, "MapPairs": nil}},
		),
	},
	{
		name: "lambda without signature",
		code: "{ echo}",
		node: &Primary{},
		want: ast{"Primary", fs{
			"Type":  Lambda,
			"Chunk": "echo",
		}},
	},
	{
		name: "new-style lambda with arguments and options",
		code: "{|a b &k=v| echo}",
		node: &Primary{},
		want: ast{"Primary", fs{
			"Type":     Lambda,
			"Elements": []string{"a", "b"},
			"MapPairs": []string{"&k=v"},
			"Chunk":    " echo",
		}},
	},
	{
		name: "output capture",
		code: "a () (b;c) (c\nd)",
		node: &Chunk{},
		want: a(
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
		),
	},
	{
		name: "exception capture",
		code: "a ?() ?(b;c)",
		node: &Chunk{},
		want: a(
			ast{"Compound/Indexing/Primary", fs{
				"Type": ExceptionCapture, "Chunk": ""}},
			ast{"Compound/Indexing/Primary", fs{
				"Type": ExceptionCapture, "Chunk": "b;c",
			}}),
	},
	{
		name: "braced list",
		code: "{,a,c\ng\n}",
		node: &Primary{},
		want: ast{"Primary", fs{
			"Type":   Braced,
			"Braced": []string{"", "a", "c", "g", ""}}},
	},
	{
		name: "tilde",
		code: "~xiaq/go",
		node: &Compound{},
		want: ast{"Compound", fs{
			"Indexings": []ast{
				{"Indexing/Primary", fs{"Type": Tilde, "Value": "~"}},
				{"Indexing/Primary", fs{"Type": Bareword, "Value": "xiaq/go"}},
			},
		}},
	},
	{
		name: "tilde and wildcard",
		code: "~xiaq/*.go",
		node: &Compound{},
		want: ast{"Compound", fs{
			"Indexings": []ast{
				{"Indexing/Primary", fs{"Type": Tilde, "Value": "~"}},
				{"Indexing/Primary", fs{"Type": Bareword, "Value": "xiaq/"}},
				{"Indexing/Primary", fs{"Type": Wildcard, "Value": "*"}},
				{"Indexing/Primary", fs{"Type": Bareword, "Value": ".go"}},
			},
		}},
	},

	{
		name:         "unterminated single-quoted string",
		code:         "'a",
		node:         &Chunk{},
		wantErrAtEnd: true,
		wantErrMsg:   "string not terminated",
	},
	{
		name:         "unterminated double-quoted string",
		code:         `"a`,
		node:         &Chunk{},
		wantErrAtEnd: true,
		wantErrMsg:   "string not terminated",
	},
	{
		name:        "invalid control sequence",
		code:        `a "\^` + "\t",
		node:        &Chunk{},
		wantErrPart: "\t",
		wantErrMsg:  "invalid control sequence, should be a codepoint between 0x3F and 0x5F",
	},
	{
		name:        "invalid hex escape sequence",
		code:        `a "\xQQ"`,
		node:        &Chunk{},
		wantErrPart: "Q",
		wantErrMsg:  "invalid escape sequence, should be hex digit",
	},
	{
		name:        "invalid octal escape sequence",
		code:        `a "\1ab"`,
		node:        &Chunk{},
		wantErrPart: "a",
		wantErrMsg:  "invalid escape sequence, should be octal digit",
	},
	{
		name:        "overflow in octal escape sequence",
		code:        `a "\400"`,
		node:        &Chunk{},
		wantErrPart: "\\400",
		wantErrMsg:  "invalid octal escape sequence, should be below 256",
	},
	{
		name:        "invalid single-char escape sequence",
		code:        `a "\i"`,
		node:        &Chunk{},
		wantErrPart: "i",
		wantErrMsg:  "invalid escape sequence",
	},
	{
		name:         "unterminated variable name",
		code:         "$",
		node:         &Chunk{},
		wantErrAtEnd: true,
		wantErrMsg:   "should be variable name",
	},
	{
		name: "list-map hybrid not supported",
		code: "a [a &k=v]",
		node: &Chunk{},
		// TODO(xiaq): Add correct position information.
		wantErrAtEnd: true,
		wantErrMsg:   "cannot contain both list elements and map pairs",
	},

	// Line continuation
	{
		name: "line continuation",
		code: "a b^\nc",
		node: &Chunk{},
		want: ast{
			"Chunk/Pipeline/Form", fs{"Head": "a", "Args": []string{"b", "c"}}},
	},
	{
		name:         "unterminated line continuation",
		code:         `a ^`,
		node:         &Chunk{},
		wantErrAtEnd: true,
		wantErrMsg:   "should be newline",
	},

	// Carriage return
	{
		name: "carriage return separating pipelines",
		code: "a\rb",
		node: &Chunk{},
		want: ast{"Chunk", fs{"Pipelines": []string{"a", "b"}}},
	},
	{
		name: "carriage return + newline separating pipelines",
		code: "a\r\nb",
		node: &Chunk{},
		want: ast{"Chunk", fs{"Pipelines": []string{"a", "b"}}},
	},
	{
		name: "carriage return as whitespace padding in lambdas",
		code: "a { \rfoo\r\nbar }",
		node: &Chunk{},
		want: a(
			ast{"Compound/Indexing/Primary",
				fs{"Type": Lambda, "Chunk": "foo\r\nbar "}},
		),
	},
	{
		name: "carriage return separating elements in a lists",
		code: "a [a\rb]",
		node: &Chunk{},
		want: a(
			ast{"Compound/Indexing/Primary", fs{
				"Type":     List,
				"Elements": []string{"a", "b"}}}),
	},
	{
		name: "carriage return in line continuation",
		code: "a b^\rc",
		node: &Chunk{},
		want: ast{
			"Chunk/Pipeline/Form", fs{"Head": "a", "Args": []string{"b", "c"}}},
	},
	{
		name: "carriage return + newline as a single newline in line continuation",
		code: "a b^\r\nc",
		node: &Chunk{},
		want: ast{
			"Chunk/Pipeline/Form", fs{"Head": "a", "Args": []string{"b", "c"}}},
	},

	// Comment
	{
		name: "comments in chunks",
		code: "a#haha\nb#lala",
		node: &Chunk{},
		want: ast{
			"Chunk", fs{"Pipelines": []ast{
				{"Pipeline/Form", fs{"Head": "a"}},
				{"Pipeline/Form", fs{"Head": "b"}},
			}}},
	},
	{
		name: "comments in lists",
		code: "a [a#haha\nb]",
		node: &Chunk{},
		want: a(
			ast{"Compound/Indexing/Primary", fs{
				"Type":     List,
				"Elements": []string{"a", "b"},
			}},
		),
	},

	// Other errors
	{
		name:        "unmatched )",
		code:        ")",
		node:        &Chunk{},
		wantErrPart: ")",
		wantErrMsg:  "unexpected rune ')'",
	},
	{
		name:        "unmatched ]",
		code:        "]",
		node:        &Chunk{},
		wantErrPart: "]",
		wantErrMsg:  "unexpected rune ']'",
	},
	{
		name:        "unmatched }",
		code:        "}",
		node:        &Chunk{},
		wantErrPart: "}",
		wantErrMsg:  "unexpected rune '}'",
	},
	{
		name:         "unmatched (",
		code:         "a (",
		node:         &Chunk{},
		wantErrAtEnd: true,
		wantErrMsg:   "should be ')'",
	},
	{
		name:         "unmatched [",
		code:         "a [",
		node:         &Chunk{},
		wantErrAtEnd: true,
		wantErrMsg:   "should be ']'",
	},
	{
		name:         "unmatched {",
		code:         "a {",
		node:         &Chunk{},
		wantErrAtEnd: true,
		wantErrMsg:   "should be ',' or '}'",
	},
	{
		name:         "unmatched { in lambda",
		code:         "a { ",
		node:         &Chunk{},
		wantErrAtEnd: true,
		wantErrMsg:   "should be '}'",
	},
	{
		name:        "unmatched [ in indexing expression",
		code:        "a $a[0}",
		node:        &Chunk{},
		wantErrPart: "}",
		wantErrMsg:  "should be ']'",
	},
}

func TestParse(t *testing.T) {
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			n := test.node
			src := SourceForTest(test.code)
			err := ParseAs(src, n, Config{})
			if test.wantErrMsg == "" {
				if err != nil {
					t.Errorf("Parse(%q) returns error: %v", test.code, err)
				}
				err = checkParseTree(n)
				if err != nil {
					t.Errorf("Parse(%q) returns bad parse tree: %v", test.code, err)
					fmt.Fprintf(os.Stderr, "Parse tree of %q:\n", test.code)
					pprintParseTree(n, os.Stderr)
				}
				err = checkAST(n, test.want)
				if err != nil {
					t.Errorf("Parse(%q) returns bad AST: %v", test.code, err)
					fmt.Fprintf(os.Stderr, "AST of %q:\n", test.code)
					pprintAST(n, os.Stderr)
				}
			} else {
				if err == nil {
					t.Errorf("Parse(%q) returns no error, want error with %q",
						test.code, test.wantErrMsg)
				}
				parseError := UnpackErrors(err)[0]
				r := parseError.Context

				if errPart := test.code[r.From:r.To]; errPart != test.wantErrPart {
					t.Errorf("Parse(%q) returns error with part %q, want %q",
						test.code, errPart, test.wantErrPart)
				}
				if atEnd := r.From == len(test.code); atEnd != test.wantErrAtEnd {
					t.Errorf("Parse(%q) returns error at end = %v, want %v",
						test.code, atEnd, test.wantErrAtEnd)
				}
				if errMsg := parseError.Message; errMsg != test.wantErrMsg {
					t.Errorf("Parse(%q) returns error with message %q, want %q",
						test.code, errMsg, test.wantErrMsg)
				}
			}
		})
	}
}

func TestParse_ReturnsTreeContainingSourceFromArgument(t *testing.T) {
	src := SourceForTest("a")
	tree, _ := Parse(src, Config{})
	if tree.Source != src {
		t.Errorf("tree.Source = %v, want %v", tree.Source, src)
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, test := range testCases {
			_ = ParseAs(SourceForTest(test.code), test.node, Config{})
		}
	}
}
