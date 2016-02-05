package parse

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

type fs map[string]interface{}
type ast struct {
	name   string
	fields fs
}

func a(c ...interface{}) ast {
	return ast{"Chunk/Pipeline/Form", fs{"Head": "a", "Args": c}}
}

var goodCases = []struct {
	src string
	ast ast
}{
	// Chunk
	{"a;b|c\n;d", ast{"Chunk", fs{"Pipelines": []string{"a", "b|c", "d"}}}},
	// Empty chunk
	{"", ast{"Chunk", nil}},
	// Lots of unnecessary whitespaces
	{"  ;\n\n  ls \t ;\n", ast{"Chunk", fs{"Pipelines": []string{"ls \t "}}}},

	// Form
	{"ls x y", ast{"Chunk/Pipeline/Form", fs{
		"Head": "ls",
		"Args": []string{"x", "y"}}}},
	// Redirections
	{"a>b", ast{"Chunk/Pipeline/Form", fs{
		"Head": "a",
		"Redirs": []ast{
			ast{"Redir", fs{"Mode": Write, "Source": "b"}}},
	}}},
	// More redirections
	{"a>>b 2>b 3>&- 4>&1", ast{"Chunk/Pipeline/Form", fs{
		"Head": "a",
		"Redirs": []ast{
			ast{"Redir", fs{"Mode": Append, "Source": "b"}},
			ast{"Redir", fs{"Mode": Write, "Source": "b"}},
			ast{"Redir", fs{"Mode": Write, "SourceIsFd": true, "Source": "-"}},
			ast{"Redir", fs{"Mode": Write, "SourceIsFd": true, "Source": "1"}},
		},
	}}},
	// Exitus redirection
	{"a ?>$e", ast{"Chunk/Pipeline/Form", fs{
		"Head":        "a",
		"ExitusRedir": ast{"ExitusRedir", fs{"Dest": "$e"}},
	}}},
	// Assignments:
	{"k=v k[a][b]=v a", ast{"Chunk/Pipeline/Form", fs{
		"Assignments": []string{"k=v", "k[a][b]=v"},
		"Head":        "a"}}},

	// Compound
	{`a b"foo"$c'xyz'`, a(ast{"Compound", fs{
		"Indexings": []string{"b", `"foo"`, "$c", "'xyz'"}}})},

	// Indexing
	{"a $b[c][d][e]", a(ast{"Compound/Indexing", fs{
		"Head": "$b", "Indicies": []string{"c", "d", "e"},
	}})},

	// Primary
	//
	// Single quote
	{"a 'b'", a(ast{"Compound/Indexing/Primary", fs{
		"text": "'b'", "Type": SingleQuoted,
	}})},
	// Double quote
	{`a "b"`, a(ast{"Compound/Indexing/Primary", fs{
		"text": `"b"`, "Type": DoubleQuoted,
	}})},
	// List
	{"a [] [ ] [1] [ 2] [3 ] [ 4 5 6 7 ]", a(
		ast{"Compound/Indexing/Primary", fs{"Type": List}},
		ast{"Compound/Indexing/Primary", fs{"Type": List}},
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
	{"a [&k v] [ &k v] [&k v ] [ &k v ] [&a b &c d &e f]", a(
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
		ast{"Compound/Indexing/Primary", fs{"Type": Map, "text": "[&]"}},
		ast{"Compound/Indexing/Primary", fs{"Type": Map, "text": "[ &]"}},
		ast{"Compound/Indexing/Primary", fs{"Type": Map, "text": "[& ]"}},
		ast{"Compound/Indexing/Primary", fs{"Type": Map, "text": "[ & ]"}},
	)},
	// Lambda
	{"a []{} [ ]{ } []{ echo 233 } [ $x $y ]{puts $x $y} { put $1}", a(
		ast{"Compound/Indexing/Primary", fs{
			"Type": Lambda,
		}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": Lambda,
		}},
		ast{"Compound/Indexing/Primary", fs{
			"Type":  Lambda,
			"Chunk": " echo 233 ",
		}},
		ast{"Compound/Indexing/Primary", fs{
			"Type":  Lambda,
			"List":  "$x $y ",
			"Chunk": "puts $x $y",
		}},
		ast{"Compound/Indexing/Primary", fs{
			"Type":  Lambda,
			"Chunk": " put $1",
		}},
	)},
	// Output capture
	{"a () (b;c)", a(
		ast{"Compound/Indexing/Primary", fs{"Type": OutputCapture}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": OutputCapture, "Chunk": "b;c",
		}})},
	// Output capture with backquotes
	{"a `` `b;c` `e>f`", a("``", "`b;c`", "`e>f`")},
	// Backquotes may be nested with unclosed parens and braces
	{"a `a (b `c`)` `d [`e`]`", a("`a (b `c`)`", "`d [`e`]`")},
	// Exitus capture
	{"a ?() ?(b;c)", a(
		ast{"Compound/Indexing/Primary", fs{"Type": ErrorCapture}},
		ast{"Compound/Indexing/Primary", fs{
			"Type": ErrorCapture, "Chunk": "b;c",
		}})},
	// Braced
	{"a {a,c-f}", a(
		ast{"Compound/Indexing/Primary", fs{
			"Type":    Braced,
			"Braced":  []string{"a", "c", "f"},
			"IsRange": []bool{false, true}}})},
}

func checkParseTree(n Node) error {
	children := n.Children()
	if len(children) == 0 {
		return nil
	}

	// check parent pointers
	for i, ch := range children {
		if ch.Parent() != n {
			return fmt.Errorf("parent of child %d (%s) is wrong: %s", i, summary(ch), summary(n))
		}
	}

	// check for possible gaps
	if children[0].Begin() != n.Begin() {
		return fmt.Errorf("gap between node and first child: %s", summary(n))
	}
	nch := len(children)
	if children[nch-1].End() != n.End() {
		return fmt.Errorf("gap between node and last child: %s", summary(n))
	}
	for i := 0; i < nch-1; i++ {
		if children[i].End() != children[i+1].Begin() {
			return fmt.Errorf("gap beteen child %d and %d of: %s", i, i+1, summary(n))
		}
	}

	// check children recursively
	for _, ch := range n.Children() {
		err := checkParseTree(ch)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkNode(got Node, want interface{}) error {
	switch want := want.(type) {
	case string:
		text := got.SourceText()
		if want != text {
			return fmt.Errorf("want %q, got %q (%s)", want, text, summary(got))
		}
		return nil
	case ast:
		return checkAST(got, want)
	default:
		panic(fmt.Sprintf("bad want type %T (%s)", want, summary(got)))
	}
}

var nodeType = reflect.TypeOf((*Node)(nil)).Elem()

func checkAny(got interface{}, want interface{}, ctx string) error {
	if got, ok := got.(Node); ok {
		// A Node
		return checkNode(got.(Node), want)
	}
	tgot := reflect.TypeOf(got)
	if tgot.Kind() == reflect.Slice && tgot.Elem().Implements(nodeType) {
		// A slice of Nodes
		vgot := reflect.ValueOf(got)
		vwant := reflect.ValueOf(want)
		if vgot.Len() != vwant.Len() {
			return fmt.Errorf("want %d, got %d (%s)", vwant.Len(), vgot.Len(), ctx)
		}
		for i := 0; i < vgot.Len(); i++ {
			err := checkNode(vgot.Index(i).Interface().(Node),
				vwant.Index(i).Interface())
			if err != nil {
				return err
			}
		}
		return nil
	}

	if !reflect.DeepEqual(want, got) {
		return fmt.Errorf("want %v, got %v (%s)", want, got, ctx)
	}
	return nil
}

func checkAST(n Node, want ast) error {
	// TODO: Check fields present in struct but not in ast
	wantnames := strings.Split(want.name, "/")
	// Check coalesced levels
	for i, wantname := range wantnames {
		name := reflect.TypeOf(n).Elem().Name()
		if wantname != name {
			return fmt.Errorf("want %s, got %s (%s)", wantname, name, summary(n))
		}
		if i == len(wantnames)-1 {
			break
		}
		fields := n.Children()
		if len(fields) != 1 {
			return fmt.Errorf("want exactly 1 child, got %d (%s)", len(fields), summary(n))
		}
		n = fields[0]
	}

	if want.fields == nil && len(n.Children()) != 0 {
		return fmt.Errorf("want leaf, got inner node (%s)", summary(n))
	}
	nv := reflect.ValueOf(n).Elem()

	for fieldname, wantfield := range want.fields {
		if fieldname == "text" {
			if n.SourceText() != wantfield.(string) {
				return fmt.Errorf("want %q, got %q (%s)", wantfield, n.SourceText())
			}
		} else {
			fv := nv.FieldByName(fieldname)
			err := checkAny(fv.Interface(), wantfield, summary(n))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func TestParse(t *testing.T) {
	for _, tc := range goodCases {
		bn, err := Parse("test", tc.src)
		if err != nil {
			t.Errorf("Parse(%q) returns error: %v", tc.src, err)
		}
		err = checkParseTree(bn)
		if err != nil {
			t.Errorf("Parse(%q) returns bad parse tree: %v", tc.src, err)
			fmt.Fprintf(os.Stderr, "Parse tree of %q:\n", tc.src)
			pprintParseTree(bn, os.Stderr)
		}
		err = checkAST(bn, tc.ast)
		if err != nil {
			t.Errorf("Parse(%q) returns bad AST: %v", tc.src, err)
			fmt.Fprintf(os.Stderr, "AST of %q:\n", tc.src)
			pprintAST(bn, os.Stderr)
		}
	}
}

// TODO: test error reporting
