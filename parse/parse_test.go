package parse

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/elves/elvish/util"
)

func compoundOfOnePrimary(p *PrimaryNode) *CompoundNode {
	return newCompound(p.Pos, NoSigil, &SubscriptNode{p.Pos, p, nil})
}

func compoundOfBare(p Pos, s string) *CompoundNode {
	return compoundOfOnePrimary(
		&PrimaryNode{p, StringPrimary, newString(p, s, s)})
}

// formWithRedir returns the expected FormNode of a "a" command followed by
// status and output redirections.
func formWithRedir(sr string, rs ...Redir) *FormNode {
	return &FormNode{0, compoundOfBare(0, "a"), newSpaced(1), rs, sr}
}

// formWithOnePrimary returns the expected FormNode of a "a" command followed
// by exactly one primary expression.
func formWithOnePrimary(p *PrimaryNode) *FormNode {
	return &FormNode{0,
		compoundOfBare(0, "a"),
		newSpaced(2, compoundOfOnePrimary(p)), nil, ""}
}

func chunkOfOneForm(f *FormNode) *ChunkNode {
	return newChunk(f.Pos, newPipeline(f.Pos, f))
}

func chunkOfFormWithRedir(sr string, rs ...Redir) *ChunkNode {
	return chunkOfOneForm(formWithRedir(sr, rs...))
}

func chunkOfFormWithOnePrimary(p *PrimaryNode) *ChunkNode {
	return chunkOfOneForm(formWithOnePrimary(p))
}

var parseTests = []struct {
	in     string
	wanted Node
}{
	// Empty chunk
	{"", newChunk(0)},

	// Command with arguments
	{"ls x y", chunkOfOneForm(
		&FormNode{0,
			compoundOfBare(0, "ls"),
			newSpaced(3,
				compoundOfBare(3, "x"),
				compoundOfBare(5, "y")),
			nil, ""})},

	// Wow... such whitespace, much unnecessary, so valid
	{"  ;\n\n  ls   ;\n", newChunk(0,
		newPipeline(7,
			&FormNode{7, compoundOfBare(7, "ls"), newSpaced(12), nil, ""}))},

	// Multiple pipelines, multiple commands
	{"a;b|c\n;d", newChunk(0,
		newPipeline(0,
			&FormNode{0,
				compoundOfBare(0, "a"),
				newSpaced(1), nil, ""}),
		newPipeline(2,
			&FormNode{2,
				compoundOfBare(2, "b"),
				newSpaced(3), nil, ""},
			&FormNode{4,
				compoundOfBare(4, "c"),
				newSpaced(5), nil, ""}),
		newPipeline(7,
			&FormNode{7,
				compoundOfBare(7, "d"),
				newSpaced(8), nil, ""}))},

	// Redirections
	//
	// Output and status redir
	{"a>b?>$c", chunkOfFormWithRedir("c",
		&FilenameRedir{
			RedirBase{1, 1}, os.O_WRONLY | os.O_CREATE,
			compoundOfBare(2, "b")})},
	// Different direction
	{"a>>b", chunkOfFormWithRedir("",
		&FilenameRedir{
			RedirBase{1, 1}, os.O_WRONLY | os.O_CREATE | os.O_APPEND,
			compoundOfBare(3, "b")})},
	// FilenameRedir with custom fd
	{"a>[2]b", chunkOfFormWithRedir("",
		&FilenameRedir{
			RedirBase{1, 2}, os.O_WRONLY | os.O_CREATE,
			compoundOfBare(5, "b")})},
	// FdRedir
	{"a>[2=33]", chunkOfFormWithRedir("", &FdRedir{RedirBase{1, 2}, 33})},
	// CloseRedir
	{"a>[2=]", chunkOfFormWithRedir("", &CloseRedir{RedirBase{1, 2}})},

	// Compound with sigil
	{"a =b$c", chunkOfOneForm(
		&FormNode{0,
			compoundOfBare(0, "a"),
			newSpaced(2,
				newCompound(2, '=',
					&SubscriptNode{3,
						&PrimaryNode{3, StringPrimary,
							newString(3, "b", "b")},
						nil},
					&SubscriptNode{4,
						&PrimaryNode{4, VariablePrimary,
							newString(5, "c", "c")},
						nil})),
			nil, ""})},

	// Subscript
	{"a $b[c]", chunkOfOneForm(
		&FormNode{0,
			compoundOfBare(0, "a"),
			newSpaced(2,
				newCompound(2, NoSigil,
					&SubscriptNode{2,
						&PrimaryNode{2, VariablePrimary,
							newString(3, "b", "b")},
						compoundOfBare(5, "c")})),
			nil, ""})},

	// Primary
	//
	// Single quote
	{"a `b`", chunkOfFormWithOnePrimary(
		&PrimaryNode{2, StringPrimary, newString(2, "`b`", "b")})},
	// Double quote
	{`a "b"`, chunkOfFormWithOnePrimary(
		&PrimaryNode{2, StringPrimary, newString(2, `"b"`, "b")})},
	// Table
	{"a [1 &k v 2 &k2 v2 3]", chunkOfFormWithOnePrimary(
		&PrimaryNode{2, TablePrimary, &TableNode{2,
			[]*CompoundNode{
				compoundOfBare(3, "1"),
				compoundOfBare(10, "2"),
				compoundOfBare(19, "3"),
			},
			[]*TablePair{
				newTablePair(compoundOfBare(6, "k"), compoundOfBare(8, "v")),
				newTablePair(compoundOfBare(13, "k2"), compoundOfBare(16, "v2")),
			}}})},
	// List
	{"a {b c}", chunkOfFormWithOnePrimary(
		&PrimaryNode{2, ListPrimary, newSpaced(3,
			compoundOfBare(3, "b"),
			compoundOfBare(5, "c"))})},
	/*
		// Closure
		{"a { b c}", nil},
		// Closure
		{"a {|b|c}", nil},
		// Channel output capture
		{"a (b c)", nil},
		// Status capture
		{"a ?(b c)", nil},
	*/
}

func TestParse(t *testing.T) {
	for i, tt := range parseTests {
		out, err := Parse(fmt.Sprintf("<test %d>", i), tt.in)
		if !reflect.DeepEqual(out, tt.wanted) || err != nil {
			t.Errorf("Parse(*, %q) =>\n(%s, %v), want\n(%s, <nil>) (up to DeepEqual)", tt.in, util.DeepPrint(out), err, util.DeepPrint(tt.wanted))
		}
	}
}

var completeTests = []struct {
	in        string
	wantedTyp ContextType
}{
	{"", CommandContext},
	{"l", CommandContext},
	{"ls ", NewArgContext},
	{"ls a", ArgContext},
	{"ls $a", ArgContext},
}

func TestComplete(t *testing.T) {
	for i, tt := range completeTests {
		out, err := Complete(fmt.Sprintf("<test %d>", i), tt.in)
		if out.Typ != tt.wantedTyp || err != nil {
			t.Errorf("Complete(*, %q) => (Context{Typ: %v, ...}, %v), want (Context{Typ: %v, ...}, <nil>)", tt.in, out.Typ, err, tt.wantedTyp)
		}
	}
}
