package complete

import (
	"strings"

	"src.elv.sh/pkg/cli/modes"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/ui"
)

// PlainItem is a simple implementation of RawItem.
type PlainItem string

func (p PlainItem) String() string { return string(p) }

func (p PlainItem) HasPrefix(seed string) bool { return strings.HasPrefix(string(p), seed) }

func (p PlainItem) Cook(q parse.PrimaryType) modes.CompletionItem {
	s := string(p)
	quoted, _ := parse.QuoteAs(s, q)
	return modes.CompletionItem{ToInsert: quoted, ToShow: s}
}

// ExternalCommand implements RawItem for external commands. On Windows, this
// will be case insensitive
type ExternalCommand struct {
	Name         string
	IsNamespaced bool
}

func (e ExternalCommand) String() string {
	if e.IsNamespaced {
		return "e:" + e.Name
	} else {
		return e.Name
	}
}

func (e ExternalCommand) HasPrefix(seed string) bool {
	if IsCaseInsensitiveOs {
		if e.IsNamespaced {
			return strings.HasPrefix(seed, "e:") &&
				strings.HasPrefix(strings.ToLower(e.Name), strings.ToLower(seed[2:]))
		} else {
			return strings.HasPrefix(strings.ToLower(e.Name), strings.ToLower(seed))
		}
	} else {
		if e.IsNamespaced {
			return strings.HasPrefix(seed, "e:") &&
				strings.HasPrefix(e.Name, seed[2:])
		} else {
			return strings.HasPrefix(e.Name, seed)
		}
	}
}

func (e ExternalCommand) Cook(q parse.PrimaryType) modes.CompletionItem {
	s := e.String()
	quoted, _ := parse.QuoteAs(s, q)
	return modes.CompletionItem{ToInsert: quoted, ToShow: s}
}

// noQuoteItem is a RawItem implementation that does not quote when cooked. This
// type is not exposed, since argument generators never need this.
type noQuoteItem string

func (nq noQuoteItem) String() string { return string(nq) }

func (nq noQuoteItem) HasPrefix(seed string) bool { return strings.HasPrefix(string(nq), seed) }

func (nq noQuoteItem) Cook(parse.PrimaryType) modes.CompletionItem {
	s := string(nq)
	return modes.CompletionItem{ToInsert: s, ToShow: s}
}

// ComplexItem is an implementation of RawItem that offers customization options.
type ComplexItem struct {
	Stem            string   // Used in the code and the menu.
	CaseInsensitive bool     // Whether to also match casing
	CodeSuffix      string   // Appended to the code.
	Display         string   // How the item is displayed. If empty, defaults to Stem.
	DisplayStyle    ui.Style // Use for displaying.
}

func (c ComplexItem) String() string { return c.Stem }

func (c ComplexItem) HasPrefix(seed string) bool {
	if c.CaseInsensitive {
		return strings.HasPrefix(strings.ToLower(c.Stem), strings.ToLower(seed))
	} else {
		return strings.HasPrefix(c.Stem, seed)
	}
}

func (c ComplexItem) Cook(q parse.PrimaryType) modes.CompletionItem {
	quoted, _ := parse.QuoteAs(c.Stem, q)
	display := c.Display
	if display == "" {
		display = c.Stem
	}
	return modes.CompletionItem{
		ToInsert:  quoted + c.CodeSuffix,
		ToShow:    display,
		ShowStyle: c.DisplayStyle,
	}
}
