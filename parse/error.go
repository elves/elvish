package parse

import (
	"bytes"
	"fmt"

	"github.com/elves/elvish/diag"
)

// ErrorEntry represents one parse error.
type ErrorEntry struct {
	Message string
	Context diag.SourceRange
}

// Error stores multiple ErrorEntry's and can pretty print them.
type Error struct {
	Entries []*ErrorEntry
}

func (pe *Error) Add(msg string, ctx *diag.SourceRange) {
	pe.Entries = append(pe.Entries, &ErrorEntry{msg, *ctx})
}

func (pe *Error) Error() string {
	switch len(pe.Entries) {
	case 0:
		return "no parse error"
	case 1:
		e := pe.Entries[0]
		return fmt.Sprintf("parse error: %d-%d in %s: %s",
			e.Context.Begin, e.Context.End, e.Context.Name, e.Message)
	default:
		buf := new(bytes.Buffer)
		// Contexts of parse error entries all have the same name
		fmt.Fprintf(buf, "multiple parse errors in %s: ", pe.Entries[0].Context.Name)
		for i, e := range pe.Entries {
			if i > 0 {
				fmt.Fprint(buf, "; ")
			}
			fmt.Fprintf(buf, "%d-%d: %s", e.Context.Begin, e.Context.End, e.Message)
		}
		return buf.String()
	}
}

func (pe *Error) PPrint(indent string) string {
	buf := new(bytes.Buffer)

	switch len(pe.Entries) {
	case 0:
		return "no parse error"
	case 1:
		e := pe.Entries[0]
		fmt.Fprintf(buf, "Parse error: \033[31;1m%s\033[m\n", e.Message)
		buf.WriteString(e.Context.PPrintCompact(indent + "  "))
	default:
		fmt.Fprint(buf, "Multiple parse errors:")
		for _, e := range pe.Entries {
			buf.WriteString("\n" + indent + "  ")
			fmt.Fprintf(buf, "\033[31;1m%s\033[m\n", e.Message)
			buf.WriteString(indent + "    ")
			buf.WriteString(e.Context.PPrint(indent + "      "))
		}
	}

	return buf.String()
}
