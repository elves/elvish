package complete

import (
	"src.elv.sh/pkg/cli/modes"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/ui"
)

// PlainItem is a simple implementation of RawItem.
type PlainItem string

func (p PlainItem) String() string { return string(p) }

func (p PlainItem) Cook(q parse.PrimaryType) modes.CompletionItem {
	s := string(p)
	quoted, _ := parse.QuoteAs(s, q)
	return modes.CompletionItem{ToInsert: quoted, ToShow: ui.T(s)}
}

// noQuoteItem is a RawItem implementation that does not quote when cooked. This
// type is not exposed, since argument generators never need this.
type noQuoteItem string

func (nq noQuoteItem) String() string { return string(nq) }

func (nq noQuoteItem) Cook(parse.PrimaryType) modes.CompletionItem {
	s := string(nq)
	return modes.CompletionItem{ToInsert: s, ToShow: ui.T(s)}
}

// ComplexItem is an implementation of RawItem that offers customization options.
type ComplexItem struct {
	Stem       string  // Used in the code and the menu.
	CodeSuffix string  // Appended to the code.
	Display    ui.Text // How the item is displayed. If empty, defaults to ui.T(Stem).
}

func (c ComplexItem) String() string { return c.Stem }

func (c ComplexItem) Cook(q parse.PrimaryType) modes.CompletionItem {
	quoted, _ := parse.QuoteAs(c.Stem, q)
	display := c.Display
	if display == nil {
		display = ui.T(c.Stem)
	}
	return modes.CompletionItem{
		ToInsert: quoted + c.CodeSuffix,
		ToShow:   display,
	}
}
