package parse

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/elves/elvish/util"
)

func compoundOfOnePrimary(p *Primary) *Compound {
	return newCompound(p.Pos, NoSigil, &Subscript{p.Pos, p, nil})
}

func compoundOfBare(p Pos, s string) *Compound {
	return compoundOfOnePrimary(
		&Primary{p, StringPrimary, newString(p, s, s)})
}

// formWithRedir returns the expected FormNode of a "a" command followed by
// status and output redirections.
func formWithRedir(sr string, rs ...Redir) *Form {
	return &Form{0, compoundOfBare(0, "a"), newSpaced(1), rs, sr}
}

// formWithOnePrimary returns the expected FormNode of a "a" command followed
// by exactly one primary expression.
func formWithOnePrimary(p *Primary) *Form {
	return &Form{0,
		compoundOfBare(0, "a"),
		newSpaced(2, compoundOfOnePrimary(p)), nil, ""}
}

func chunkOfOneForm(f *Form) *Chunk {
	return newChunk(f.Pos, newPipeline(f.Pos, f))
}

func chunkOfFormWithRedir(sr string, rs ...Redir) *Chunk {
	return chunkOfOneForm(formWithRedir(sr, rs...))
}

func chunkOfFormWithOnePrimary(p *Primary) *Chunk {
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
		&Form{0,
			compoundOfBare(0, "ls"),
			newSpaced(3,
				compoundOfBare(3, "x"),
				compoundOfBare(5, "y")),
			nil, ""})},

	// Wow... such whitespace, much unnecessary, so valid
	{"  ;\n\n  ls   ;\n", newChunk(0,
		newPipeline(7,
			&Form{7, compoundOfBare(7, "ls"), newSpaced(12), nil, ""}))},

	// Multiple pipelines, multiple commands
	{"a;b|c\n;d", newChunk(0,
		newPipeline(0,
			&Form{0,
				compoundOfBare(0, "a"),
				newSpaced(1), nil, ""}),
		newPipeline(2,
			&Form{2,
				compoundOfBare(2, "b"),
				newSpaced(3), nil, ""},
			&Form{4,
				compoundOfBare(4, "c"),
				newSpaced(5), nil, ""}),
		newPipeline(7,
			&Form{7,
				compoundOfBare(7, "d"),
				newSpaced(8), nil, ""}))},

	// Redirections
	//
	// Output and status redir
	{"a>b?>$c", chunkOfFormWithRedir("c",
		&FilenameRedir{
			redirBase{1, 1}, os.O_WRONLY | os.O_CREATE,
			compoundOfBare(2, "b")})},
	// Different direction
	{"a>>b", chunkOfFormWithRedir("",
		&FilenameRedir{
			redirBase{1, 1}, os.O_WRONLY | os.O_CREATE | os.O_APPEND,
			compoundOfBare(3, "b")})},
	// FilenameRedir with custom fd
	{"a>[2]b", chunkOfFormWithRedir("",
		&FilenameRedir{
			redirBase{1, 2}, os.O_WRONLY | os.O_CREATE,
			compoundOfBare(5, "b")})},
	// FdRedir
	{"a>[2=33]", chunkOfFormWithRedir("", &FdRedir{redirBase{1, 2}, 33})},
	// CloseRedir
	{"a>[2=]", chunkOfFormWithRedir("", &CloseRedir{redirBase{1, 2}})},

	// Compound with sigil
	{"a !b$c", chunkOfOneForm(
		&Form{0,
			compoundOfBare(0, "a"),
			newSpaced(2,
				newCompound(2, '!',
					&Subscript{3,
						&Primary{3, StringPrimary,
							newString(3, "b", "b")},
						nil},
					&Subscript{4,
						&Primary{4, VariablePrimary,
							newString(5, "c", "c")},
						nil})),
			nil, ""})},

	// Subscript
	{"a $b[c]", chunkOfOneForm(
		&Form{0,
			compoundOfBare(0, "a"),
			newSpaced(2,
				newCompound(2, NoSigil,
					&Subscript{2,
						&Primary{2, VariablePrimary,
							newString(3, "b", "b")},
						compoundOfBare(5, "c")})),
			nil, ""})},

	// Primary
	//
	// Single quote
	{"a `b`", chunkOfFormWithOnePrimary(
		&Primary{2, StringPrimary, newString(2, "`b`", "b")})},
	// Double quote
	{`a "b"`, chunkOfFormWithOnePrimary(
		&Primary{2, StringPrimary, newString(2, `"b"`, "b")})},
	// Table
	{"a [1 &k v 2 &k2 v2 3]", chunkOfFormWithOnePrimary(
		&Primary{2, TablePrimary, &Table{2,
			[]*Compound{
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
		&Primary{2, ListPrimary, newSpaced(3,
			compoundOfBare(3, "b"),
			compoundOfBare(5, "c"))})},
	// Closure: empty
	{"a { }", chunkOfFormWithOnePrimary(
		&Primary{2, ClosurePrimary, &Closure{3, nil, newChunk(4)}})},
	// Closure: parameterless
	{"a { b c}", chunkOfFormWithOnePrimary(
		&Primary{2, ClosurePrimary, &Closure{3, nil, chunkOfOneForm(
			&Form{4,
				compoundOfBare(4, "b"),
				newSpaced(6, compoundOfBare(6, "c")),
				nil, ""})}})},
	// Closure: simple with parameters
	{"a {|b|c}", chunkOfFormWithOnePrimary(
		&Primary{2, ClosurePrimary, &Closure{3,
			newSpaced(4, compoundOfBare(4, "b")),
			chunkOfOneForm(
				&Form{6,
					compoundOfBare(6, "c"), newSpaced(7), nil, ""})}})},
	// Channel output capture
	{"a (b c)", chunkOfFormWithOnePrimary(
		&Primary{2, ChanCapturePrimary,
			newPipeline(3,
				&Form{3,
					compoundOfBare(3, "b"),
					newSpaced(5, compoundOfBare(5, "c")),
					nil, ""})})},
	// Status capture
	{"a ?(b c)", chunkOfFormWithOnePrimary(
		&Primary{2, StatusCapturePrimary,
			newPipeline(4,
				&Form{4,
					compoundOfBare(4, "b"),
					newSpaced(6, compoundOfBare(6, "c")),
					nil, ""})})},
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
