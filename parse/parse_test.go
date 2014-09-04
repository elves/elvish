package parse

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/xiaq/elvish/util"
)

var parseTests = []struct {
	in     string
	wanted Node
}{
	{"", newChunk(0)},
	{"ls", newChunk( // chunk
		0, newPipeline( // pipeline
			0, &FormNode{ // form
				0, newTerm( // term
					0, &PrimaryNode{
						0, StringPrimary, newString(0, "ls", "ls")}),
				newTermList(2), nil, ""}))},
}

func TestParse(t *testing.T) {
	for i, tt := range parseTests {
		out, err := Parse(fmt.Sprintf("<test %d>", i), tt.in)
		if !reflect.DeepEqual(out, tt.wanted) || err != nil {
			t.Errorf("Parse(*, %q) =>\n(%s, %v), want\n(%s, nil) (up to DeepEqual)", tt.in, util.DeepPrint(out), err, util.DeepPrint(tt.wanted))
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
			t.Errorf("Complete(*, %q) => (Context{Typ: %v, ...}, %v), want (Context{Typ: %v, ...}, nil)", tt.in, out.Typ, err, tt.wantedTyp)
		}
	}
}
