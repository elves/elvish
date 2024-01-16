package diag

import (
	"fmt"
	"strings"

	"src.elv.sh/pkg/strutil"
)

// Error represents an error with context that can be showed.
type Error[T ErrorTag] struct {
	Message string
	Context Context
}

// ErrorTag is used to parameterize [Error] into different concrete types. The
// ErrorTag method is called with a zero receiver, and its return value is used
// in [Error.Error] and [Error.Show].
type ErrorTag interface {
	ErrorTag() string
}

// RangeError combines error with [Ranger].
type RangeError interface {
	error
	Ranger
}

// Error returns a plain text representation of the error.
func (e *Error[T]) Error() string {
	return errorTag[T]() + ": " + e.errorNoType()
}

func (e *Error[T]) errorNoType() string {
	return e.Context.describeRange() + ": " + e.Message
}

// Range returns the range of the error.
func (e *Error[T]) Range() Ranging {
	return e.Context.Range()
}

var (
	messageStart = "\033[31;1m"
	messageEnd   = "\033[m"
)

// Show shows the error.
func (e *Error[T]) Show(indent string) string {
	return errorTagTitle[T]() + ": " + e.showNoType(indent)
}

func (e *Error[T]) showNoType(indent string) string {
	indent += "  "
	return messageStart + e.Message + messageEnd +
		"\n" + indent + e.Context.Show(indent)
}

// PackErrors packs multiple instances of [Error] with the same tag into one
// error:
//
//   - If called with no errors, it returns nil.
//
//   - If called with one error, it returns that error itself.
//
//   - If called with more than one [Error], it returns an error that combines
//     all of them. The returned error also implements [Shower], and its Error
//     and Show methods only print the tag once.
func PackErrors[T ErrorTag](errs []*Error[T]) error {
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return append(multiError[T](nil), errs...)
	}
}

// UnpackErrors returns the constituent [Error] instances in an error if it is
// built from [PackErrors]. Otherwise it returns nil.
func UnpackErrors[T ErrorTag](err error) []*Error[T] {
	switch err := err.(type) {
	case *Error[T]:
		return []*Error[T]{err}
	case multiError[T]:
		return append([]*Error[T](nil), err...)
	default:
		return nil
	}
}

type multiError[T ErrorTag] []*Error[T]

func (err multiError[T]) Error() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "multiple %s: ", errorTagPlural[T]())
	for i, e := range err {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(e.errorNoType())
	}
	return sb.String()
}

func (err multiError[T]) Show(indent string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Multiple %s:", errorTagPlural[T]())
	indent += "  "
	for _, e := range err {
		sb.WriteString("\n" + indent)
		sb.WriteString(e.showNoType(indent))
	}
	return sb.String()
}

func errorTag[T ErrorTag]() string {
	var t T
	return t.ErrorTag()
}

// We don't have any error tags with an irregular plural yet. When we do, we can
// let ErrorTag optionally implement interface{ ErrorTagPlural() } and use that
// when available.
func errorTagPlural[T ErrorTag]() string { return errorTag[T]() + "s" }

func errorTagTitle[T ErrorTag]() string { return strutil.Title(errorTag[T]()) }
