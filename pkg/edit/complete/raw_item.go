package complete

import (
	"github.com/elves/elvish/pkg/cli/addons/completion"
	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/ui"
)

// PlainItem is a simple implementation of RawItem.
type PlainItem string

func (p PlainItem) String() string { return string(p) }

func (p PlainItem) Cook(q parse.PrimaryType) completion.Item {
	s := string(p)
	quoted, _ := parse.QuoteAs(s, q)
	return completion.Item{ToInsert: quoted, ToShow: s}
}

// noQuoteItem is a RawItem implementation that does not quote when cooked. This
// type is not exposed, since argument generators never need this.
type noQuoteItem string

func (nq noQuoteItem) String() string { return string(nq) }

func (nq noQuoteItem) Cook(parse.PrimaryType) completion.Item {
	s := string(nq)
	return completion.Item{ToInsert: s, ToShow: s}
}

// ComplexItem is an implementation of RawItem that offers customization options.
type ComplexItem struct {
	Stem          string   // Used in the code and the menu.
	CodeSuffix    string   // Appended to the code.
	DisplaySuffix string   // Appended to the display.
	DisplayStyle  ui.Style // Use for displaying.
}

func (c ComplexItem) String() string { return c.Stem }

func (c ComplexItem) Cook(q parse.PrimaryType) completion.Item {
	quoted, _ := parse.QuoteAs(c.Stem, q)
	return completion.Item{
		ToInsert:  quoted + c.CodeSuffix,
		ToShow:    c.Stem + c.DisplaySuffix,
		ShowStyle: c.DisplayStyle,
	}
}
