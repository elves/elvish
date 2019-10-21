package complete

import (
	"github.com/elves/elvish/cli/addons/completion"
	"github.com/elves/elvish/parse"
)

// plainItem is a minimal implementation of rawItem.
type plainItem string

func (p plainItem) String() string { return string(p) }

func (p plainItem) Cook(q parse.PrimaryType) completion.Item {
	s := string(p)
	quoted, _ := parse.QuoteAs(s, q)
	return completion.Item{ToInsert: quoted, ToShow: s}
}

// noQuoteItem is a rawItem that does not quote when cooked.
type noQuoteItem string

func (nq noQuoteItem) String() string { return string(nq) }

func (nq noQuoteItem) Cook(parse.PrimaryType) completion.Item {
	s := string(nq)
	return completion.Item{ToInsert: s, ToShow: s}
}

// complexItem is an implementation of rawItem that offers ustomization options.
type complexItem struct {
	stem          string // Used in the code and the menu.
	codeSuffix    string // Appended to the code.
	displaySuffix string // Appended to the display.
}

func (c *complexItem) String() string { return c.stem }

func (c *complexItem) Cook(q parse.PrimaryType) completion.Item {
	quoted, _ := parse.QuoteAs(c.stem, q)
	return completion.Item{
		ToInsert: quoted + c.codeSuffix,
		ToShow:   c.stem + c.displaySuffix,
	}
}
