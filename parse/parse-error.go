package parse

import (
	"bytes"
	"fmt"

	"github.com/elves/elvish/util"
)

// ParseErrorEntry represents one parse error.
type ParseErrorEntry struct {
	Message string
	Context util.SourceContext
}

// ParseError stores multiple ParseErrorEntry's and can pretty print them.
type ParseError struct {
	Entries []*ParseErrorEntry
}

func (pe *ParseError) Add(msg string, ctx util.SourceContext) {
	pe.Entries = append(pe.Entries, &ParseErrorEntry{msg, ctx})
}

func (pe *ParseError) Error() string {
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

func (pe *ParseError) Pprint(indent string) string {
	buf := new(bytes.Buffer)

	switch len(pe.Entries) {
	case 0:
		return "no parse error"
	case 1:
		e := pe.Entries[0]
		fmt.Fprintf(buf, "Parse error: \033[31;1m%s\033[m\n", e.Message)
		buf.WriteString(indent + "  ")
		e.Context.Pprint(buf, indent+"    ")
	default:
		fmt.Fprint(buf, "Multiple parse errors:")
		for _, e := range pe.Entries {
			buf.WriteString("\n" + indent + "  ")
			fmt.Fprintf(buf, "\033[31;1m%s\033[m\n", e.Message)
			buf.WriteString(indent + "    ")
			e.Context.Pprint(buf, indent+"      ")
		}
	}

	return buf.String()
}
