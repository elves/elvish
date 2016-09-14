package edit

import (
	"fmt"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

type candidate struct {
	text    string
	display styled
	suffix  string
}

var _ eval.Value = &candidate{}

func (c *candidate) Kind() string {
	return "map"
}

func (c *candidate) Repr(int) string {
	return fmt.Sprintf("(le:candidate %s %s %s)", parse.Quote(c.text), c.display.Repr(eval.NoPretty), parse.Quote(c.suffix))
}
