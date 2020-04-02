package parse

import (
	"fmt"
	"strings"

	"github.com/elves/elvish/pkg/diag"
)

const parseErrorType = "parse error"

// MultiError stores multiple Error's and can pretty print them.
type MultiError struct {
	Entries []*diag.Error
}

var _ diag.Shower = &MultiError{}

func (me *MultiError) add(msg string, ctx *diag.Context) {
	err := &diag.Error{Type: parseErrorType, Message: msg, Context: *ctx}
	me.Entries = append(me.Entries, err)
}

// Error returns a string representation of the error.
func (me *MultiError) Error() string {
	switch len(me.Entries) {
	case 0:
		return "no parse error"
	case 1:
		return me.Entries[0].Error()
	default:
		sb := new(strings.Builder)
		// Contexts of parse error entries all have the same name
		fmt.Fprintf(sb, "multiple parse errors in %s: ", me.Entries[0].Context.Name)
		for i, e := range me.Entries {
			if i > 0 {
				fmt.Fprint(sb, "; ")
			}
			fmt.Fprintf(sb, "%d-%d: %s", e.Context.From, e.Context.To, e.Message)
		}
		return sb.String()
	}
}

// Show shows the error.
func (me *MultiError) Show(indent string) string {
	switch len(me.Entries) {
	case 0:
		return "no parse error"
	case 1:
		return me.Entries[0].Show(indent)
	default:
		sb := new(strings.Builder)
		fmt.Fprint(sb, "Multiple parse errors:")
		for _, e := range me.Entries {
			sb.WriteString("\n" + indent + "  ")
			fmt.Fprintf(sb, "\033[31;1m%s\033[m\n", e.Message)
			sb.WriteString(indent + "    ")
			sb.WriteString(e.Context.Show(indent + "      "))
		}
		return sb.String()
	}
}
