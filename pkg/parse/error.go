package parse

import (
	"fmt"
	"strings"

	"src.elv.sh/pkg/diag"
)

const parseErrorType = "parse error"

// Error stores multiple underlying parse errors, and can pretty print them.
type Error struct {
	Entries []*diag.Error
}

var _ diag.Shower = &Error{}

// GetError returns an *Error if the given error has dynamic type *Error, i.e.
// is returned by one of the Parse functions. Otherwise it returns nil.
func GetError(e error) *Error {
	if er, ok := e.(*Error); ok {
		return er
	}
	return nil
}

func (er *Error) add(msg string, ctx *diag.Context) {
	err := &diag.Error{Type: parseErrorType, Message: msg, Context: *ctx}
	er.Entries = append(er.Entries, err)
}

// Error returns a string representation of the error.
func (er *Error) Error() string {
	switch len(er.Entries) {
	case 0:
		return "no parse error"
	case 1:
		return er.Entries[0].Error()
	default:
		sb := new(strings.Builder)
		// Contexts of parse error entries all have the same name
		fmt.Fprintf(sb, "multiple parse errors in %s: ", er.Entries[0].Context.Name)
		for i, e := range er.Entries {
			if i > 0 {
				fmt.Fprint(sb, "; ")
			}
			fmt.Fprintf(sb, "%d-%d: %s", e.Context.From, e.Context.To, e.Message)
		}
		return sb.String()
	}
}

// Show shows the error.
func (er *Error) Show(indent string) string {
	switch len(er.Entries) {
	case 0:
		return "no parse error"
	case 1:
		return er.Entries[0].Show(indent)
	default:
		sb := new(strings.Builder)
		fmt.Fprint(sb, "Multiple parse errors:")
		for _, e := range er.Entries {
			sb.WriteString("\n" + indent + "  ")
			fmt.Fprintf(sb, "\033[31;1m%s\033[m\n", e.Message)
			sb.WriteString(indent + "    ")
			sb.WriteString(e.Context.Show(indent + "      "))
		}
		return sb.String()
	}
}
