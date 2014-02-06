package parse

import (
	"fmt"
	"github.com/xiaq/elvish/util"
	"reflect"
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
				newList(2), nil}))},
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
	in     string
	wanted *Context
}{
	{"", &Context{
		CommandContext, nil,
		newList(0), newList(0),
		&FactorNode{0, StringFactor, newString(0, "", "")}}},
	{"l", &Context{
		CommandContext, nil,
		newList(0), newList(0),
		&FactorNode{0, StringFactor, newString(0, "l", "l")}}},
	{"ls ", &Context{
		ArgContext,
		newList(0, &FactorNode{0, StringFactor, newString(0, "ls", "ls")}),
		newList(3),
		newList(3),
		&FactorNode{3, StringFactor, newString(3, "", "")}}},
	{"ls a", &Context{
		ArgContext,
		newList(0, &FactorNode{0, StringFactor, newString(0, "ls", "ls")}),
		newList(3),
		newList(3),
		&FactorNode{3, StringFactor, newString(3, "a", "a")}}},
	{"ls $a", &Context{
		ArgContext,
		newList(0, &FactorNode{0, StringFactor, newString(0, "ls", "ls")}),
		newList(3),
		newList(3),
		&FactorNode{3, VariableFactor, newString(4, "a", "a")}}},
}

func TestComplete(t *testing.T) {
	for i, tt := range completeTests {
		out, err := Complete(fmt.Sprintf("<test %d>", i), tt.in)
		if !reflect.DeepEqual(out, tt.wanted) || err != nil {
			t.Errorf("Complete(*, %q) =>\n(%s, %v), want\n(%s, nil)", tt.in, util.DeepPrint(out), err, util.DeepPrint(tt.wanted))
		}
	}
}
