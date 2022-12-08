package diag

import (
	"fmt"
	"strings"

	"src.elv.sh/pkg/wcwidth"
)

// PackCognateErrors combines multiple instances of [Error] with with the same
// Type and Context.Name into one:
//
//   - If called with no errors, it returns nil.
//
//   - If called with one error, it returns that error itself.
//
//   - If called with more than one [Error], it returns an error that combines
//     all of them. The returned error also implements [Shower], and its Error
//     and Show methods avoid duplicating the type and context name of the
//     constituent errors.
func PackCognateErrors(errs []*Error) error {
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return append(cognateErrors(nil), errs...)
	}
}

// UnpackCognateErrors returns the constituent [Error] instances in an error and
// if it is built from [PackCognateErrors]. Otherwise it returns nil.
func UnpackCognateErrors(err error) []*Error {
	switch err := err.(type) {
	case *Error:
		return []*Error{err}
	case cognateErrors:
		return append([]*Error(nil), err...)
	default:
		return nil
	}
}

type cognateErrors []*Error

func (err cognateErrors) Error() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "multiple %ss in %s: ", err[0].Type, err[0].Context.Name)
	for i, e := range err {
		if i > 0 {
			sb.WriteString("; ")
		}
		fmt.Fprintf(&sb, "%s: %s", e.Context.culprit.describeStart(), e.Message)
	}
	return sb.String()
}

func (err cognateErrors) Show(indent string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Multiple %ss in %s:", err[0].Type, err[0].Context.Name)
	for _, e := range err {
		sb.WriteString("\n" + indent + "  ")
		sb.WriteString(messageStart + e.Message + messageEnd)
		sb.WriteString("\n" + indent + "    ")
		// This duplicates part of [Context.ShowCompact].
		desc := e.Context.culprit.describeStart() + ": "
		descIndent := strings.Repeat(" ", wcwidth.Of(desc))
		sb.WriteString(desc + e.Context.culprit.Show(indent+"  "+descIndent))
	}
	return sb.String()
}
