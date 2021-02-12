package complete

import (
	"src.elv.sh/pkg/cli/mode"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/ui"
)

// PlainItem is a simple implementation of RawItem.
type PlainItem string

func (p PlainItem) String() string { return string(p) }

func (p PlainItem) Cook(q parse.PrimaryType) mode.CompletionItem {
	s := string(p)
	quoted, _ := parse.QuoteAs(s, q)
	return mode.CompletionItem{ToInsert: quoted, ToShow: s}
}

// noQuoteItem is a RawItem implementation that does not quote when cooked. This
// type is not exposed, since argument generators never need this.
type noQuoteItem string

func (nq noQuoteItem) String() string { return string(nq) }

func (nq noQuoteItem) Cook(parse.PrimaryType) mode.CompletionItem {
	s := string(nq)
	return mode.CompletionItem{ToInsert: s, ToShow: s}
}

// ComplexItem is an implementation of RawItem that offers customization options.
type ComplexItem struct {
	Stem         string   // Used in the code and the menu.
	CodeSuffix   string   // Appended to the code.
	Display      string   // How the item is displayed. If empty, defaults to Stem.
	DisplayStyle ui.Style // Use for displaying.
}

func (c ComplexItem) String() string { return c.Stem }

func (c ComplexItem) Cook(q parse.PrimaryType) mode.CompletionItem {
	quoted, _ := parse.QuoteAs(c.Stem, q)
	display := c.Display
	if display == "" {
		display = c.Stem
	}
	return mode.CompletionItem{
		ToInsert:  quoted + c.CodeSuffix,
		ToShow:    display,
		ShowStyle: c.DisplayStyle,
	}
}
