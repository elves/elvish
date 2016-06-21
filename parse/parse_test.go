package parse

import (
	"fmt"
	"os"
	"testing"

	"github.com/elves/elvish/util"
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
	{"k=v k[a][b]=v", ast{"Chunk/Pipeline/Form", fs{
		"Assignments": []string{"k=v", "k[a][b]=v"}}}},
	// Temporary assignment.
	{"k=v k[a][b]=v a", ast{"Chunk/Pipeline/Form", fs{
		"Assignments": []string{"k=v", "k[a][b]=v"},
		"Head":        "a"}}},
	// Redirections
	{"a >b", ast{"Chunk/Pipeline/Form", fs{
		"Head": "a",
		"Redirs": []ast{
			ast{"Redir", fs{"Mode": Write, "Source": "b"}}},
	}}},
	// More redirections
	{"a >>b 2>b 3>&- 4>&1 5<c 6<>d", ast{"Chunk/Pipeline/Form", fs{
		"Head": "a",
		"Redirs": []ast{
			ast{"Redir", fs{"Mode": Append, "Source": "b"}},
			ast{"Redir", fs{"Dest": "2", "Mode": Write, "Source": "b"}},
			ast{"Redir", fs{"Dest": "3", "Mode": Write, "SourceIsFd": true, "Source": "-"}},
			ast{"Redir", fs{"Dest": "4", "Mode": Write, "SourceIsFd": true, "Source": "1"}},
			ast{"Redir", fs{"Dest": "5", "Mode": Read, "Source": "c"}},
			ast{"Redir", fs{"Dest": "6", "Mode": ReadWrite, "Source": "d"}},
		},
	}}},
	// Exitus redirection
	{"a ?>$e", ast{"Chunk/Pipeline/Form", fs{
		"Head":        "a",
		"ExitusRedir": ast{"ExitusRedir", fs{"Dest": "$e"}},
	}}},
	// TODO Names arguments.

	// Control structures.
	// if/then/fi.
	{"if true; then echo then; fi",
		ast{"Chunk/Pipeline/Form/Control", fs{
			"Kind":       IfControl,
			"Conditions": []string{" true; "},
			"Bodies":     []string{" echo then; "},
		}}},
	// if/then/else/fi.
	{"if true; then echo then; else echo else; fi",
		ast{"Chunk/Pipeline/Form/Control", fs{
			"Kind":       IfControl,
			"Conditions": []string{" true; "},
			"Bodies":     []string{" echo then; "},
			"ElseBody":   " echo else; ",
		}}},
	// if/then/elif/then/else/fi.
	{"if true; then echo then; elif true; then echo else if; else echo else; fi",
		ast{"Chunk/Pipeline/Form/Control", fs{
			"Kind":       IfControl,
			"Conditions": []string{" true; ", " true; "},
			"Bodies":     []string{" echo then; ", " echo else if; "},
			"ElseBody":   " echo else; ",
		}}},
	// while/do/done
	{"while true; do echo do; done",
		ast{"Chunk/Pipeline/Form/Control", fs{
			"Kind":      WhileControl,
			"Condition": " true; ", "Body": " echo do; "}}},
	// while/do/else/done
	{"while true; do echo do; else echo else; done",
		ast{"Chunk/Pipeline/Form/Control", fs{
			"Kind":      WhileControl,
			"Condition": " true; ",
			"Body":      " echo do; ",
			"ElseBody":  " echo else; ",
		}}},
	// for/do/done
	{"for\nx\nin\na\nb c; do echo do; done",
		ast{"Chunk/Pipeline/Form/Control", fs{
			"Kind":     ForControl,
			"Iterator": "x",
			"Array":    "\na\nb c",
			"Body":     " echo do; "}}},
	// begin/end
	{"begin echo begin; end",
		ast{"Chunk/Pipeline/Form/Control", fs{
			"Kind": BeginControl, "Body": " echo begin; "}}},

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
	// Map
	{"a [&k=v] [ &k=v] [&k=v ] [ &k=v ] [&a=b &c=d &e=f]", a(
		ast{"Compound/Indexing/Primary", fs{
			"Type":     Map,
			"MapPairs": []ast{ast{"MapPair", fs{"Key": "k", "Value": "v"}}}}},
		ast{"Compound/Indexing/Primary", fs{
			"Type":     Map,
			"MapPairs": []ast{ast{"MapPair", fs{"Key": "k", "Value": "v"}}}}},
		ast{"Compound/Indexing/Primary", fs{
			"Type":     Map,
			"MapPairs": []ast{ast{"MapPair", fs{"Key": "k", "Value": "v"}}}}},
		ast{"Compound/Indexing/Primary", fs{
			"Type":     Map,
			"MapPairs": []ast{ast{"MapPair", fs{"Key": "k", "Value": "v"}}}}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": Map,
			"MapPairs": []ast{
				ast{"MapPair", fs{"Key": "a", "Value": "b"}},
				ast{"MapPair", fs{"Key": "c", "Value": "d"}},
				ast{"MapPair", fs{"Key": "e", "Value": "f"}},
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
			"Type": ErrorCapture, "Chunk": ""}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": ErrorCapture, "Chunk": "b;c",
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
		bn, err := Parse(tc.src)
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
			return fmt.Errorf("gap beteen child %d and %d of: %s", i, i+1, summary(n))
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
	{"a &", 3}, {"a [&", 4},
	// Bogus command leader.
	{"else echo 233", 0},
}

func TestParseError(t *testing.T) {
	for _, tc := range badCases {
		_, err := Parse(tc.src)
		if err == nil {
			t.Errorf("Parse(%q) returns no error", tc.src)
			continue
		}
		posErr0 := err.(*util.Errors).Errors[0].(*util.PosError)
		if posErr0.Begin != tc.pos {
			t.Errorf("Parse(%q) first error begins at %d, want %d. Errors are:%s\n", tc.src, posErr0.Begin, tc.pos, err)
		}
	}
}
