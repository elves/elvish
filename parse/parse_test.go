package parse

import (
	"fmt"
	"testing"
)

var parseTests = []struct {
	in     string
	wanted Node
}{
	{"", newList(0)},
	{"ls", newList( // chunk
		0, newList( // pipeline
			0, &FormNode{ // form
				0, newList( // term
					0, &FactorNode{ // factor
						0, StringFactor, newString(0, "ls", "ls")}),
				newList(0), []Redir{}}))},
}

func TestParse(t *testing.T) {
	for i, tt := range parseTests {
		out, err := Parse(fmt.Sprintf("<test %d>", i), tt.in)
		if out == nil || !out.Isomorph(tt.wanted) || err != nil {
			t.Errorf("Parse(*, %q) => (%v, %v), want (%v, nil) (up to isomorphism)", tt.in, out, err, tt.wanted)
		}
	}
}

var completeTests = []struct {
	in     string
	wanted *Context
}{
	{"", &Context{CommandContext, ""}},
	{"l", &Context{CommandContext, "l"}},
	{"ls ", &Context{NewTermContext, ""}},
	{"ls a", &Context{FilenameContext, "a"}},
	{"ls $a", &Context{VariableNameContext, "a"}},
}

func TestComplete(t *testing.T) {
	for i, tt := range completeTests {
		out, err := Complete(fmt.Sprintf("<test %d>", i), tt.in)
		if out == nil || *out != *tt.wanted || err != nil {
			t.Errorf("Complete(*, %q) => (%v, %v), want (%v, nil)", tt.in, out, err, tt.wanted)
		}
	}
}
