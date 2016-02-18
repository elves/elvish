package eval

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/elves/elvish/parse"
)

// Error represents runtime errors in elvish constructs.
type Error struct {
	inner error
}

func (Error) Kind() string {
	return "error"
}

func (e Error) Repr() string {
	if e.inner == nil {
		return "$ok"
	}
	if r, ok := e.inner.(Reprer); ok {
		return r.Repr()
	}
	return "?(error " + parse.Quote(e.inner.Error()) + ")"
}

func (e Error) Bool() bool {
	return e.inner == nil
}

// Common Error values.
var (
	OK             = Error{nil}
	GenericFailure = Error{errors.New("generic failure")}
)

// multiError is multiple errors packed into one. It is used for reporting
// errors of pipelines, in which multiple forms may error.
type multiError struct {
	errors []Error
}

func (me multiError) Repr() string {
	b := new(bytes.Buffer)
	b.WriteString("?(multi-error")
	for _, e := range me.errors {
		b.WriteString(" ")
		b.WriteString(e.Repr())
	}
	b.WriteString(")")
	return b.String()
}

func (me multiError) Error() string {
	b := new(bytes.Buffer)
	b.WriteString("(")
	for i, e := range me.errors {
		if i > 0 {
			b.WriteString(" | ")
		}
		b.WriteString(e.inner.Error())
	}
	b.WriteString(")")
	return b.String()
}

func newMultiError(es ...Error) Error {
	return Error{multiError{es}}
}

// Flow is a special type of Error used for control flows.
type flow uint

// Control flows.
const (
	Return flow = iota
	Break
	Continue
)

var flowNames = [...]string{
	"return", "break", "continue",
}

func (f flow) Repr() string {
	return "?(" + f.Error() + ")"
}

func (f flow) Error() string {
	if f >= flow(len(flowNames)) {
		return fmt.Sprintf("!(BAD FLOW: %v)", f)
	}
	return flowNames[f]
}

func allok(es []Error) bool {
	for _, e := range es {
		if e.inner != nil {
			return false
		}
	}
	return true
}

// PprintError pretty prints an error. It understands specialized error types
// defined in this package.
func PprintError(e error) {
	switch e := e.(type) {
	case nil:
		fmt.Print("\033[32mok\033[m")
	case multiError:
		fmt.Print("(")
		for i, c := range e.errors {
			if i > 0 {
				fmt.Print(" | ")
			}
			PprintError(c.inner)
		}
		fmt.Print(")")
	case flow:
		fmt.Print("\033[33m" + e.Error() + "\033[m")
	default:
		fmt.Print("\033[31;1m" + e.Error() + "\033[m")
	}
}
