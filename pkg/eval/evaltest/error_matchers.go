package evaltest

import (
	"fmt"
	"reflect"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/parse"
)

type errorMatcher interface{ matchError(error) bool }

// An errorMatcher for compilation errors.
type compilationError struct {
	msgs []string
}

func (e compilationError) Error() string {
	return fmt.Sprintf("compilation errors with messages: %v", e.msgs)
}

func (e compilationError) matchError(e2 error) bool {
	errs := eval.UnpackCompilationErrors(e2)
	if len(e.msgs) != len(errs) {
		return false
	}
	for i, msg := range e.msgs {
		if msg != errs[i].Message {
			return false
		}
	}
	return true
}

// An errorMatcher for exceptions.
type exc struct {
	reason error
	stacks []string
}

func (e exc) Error() string {
	if len(e.stacks) == 0 {
		return fmt.Sprintf("exception with reason %v", e.reason)
	}
	return fmt.Sprintf("exception with reason %v and stacks %v", e.reason, e.stacks)
}

func (e exc) matchError(e2 error) bool {
	if e2, ok := e2.(eval.Exception); ok {
		return matchErr(e.reason, e2.Reason()) &&
			(len(e.stacks) == 0 ||
				reflect.DeepEqual(e.stacks, getStackTexts(e2.StackTrace())))
	}
	return false
}

func getStackTexts(tb *eval.StackTrace) []string {
	texts := []string{}
	for tb != nil {
		ctx := tb.Head
		texts = append(texts, ctx.Body)
		tb = tb.Next
	}
	return texts
}

// AnyParseError is an error that can be passed to the Case.Throws to match any
// parse error.
var AnyParseError anyParseError

type anyParseError struct{}

func (anyParseError) Error() string           { return "any parse error" }
func (anyParseError) matchError(e error) bool { return parse.UnpackErrors(e) != nil }

// ErrorWithType returns an error that can be passed to the Case.Throws to match
// any error with the same type as the argument.
func ErrorWithType(v error) error { return errWithType{v} }

// An errorMatcher for any error with the given type.
type errWithType struct{ v error }

func (e errWithType) Error() string { return fmt.Sprintf("error with type %T", e.v) }

func (e errWithType) matchError(e2 error) bool {
	return reflect.TypeOf(e.v) == reflect.TypeOf(e2)
}

// ErrorWithMessage returns an error that can be passed to Case.Throws to match
// any error with the given message.
func ErrorWithMessage(msg string) error { return errWithMessage{msg} }

// An errorMatcher for any error with the given message.
type errWithMessage struct{ msg string }

func (e errWithMessage) Error() string { return "error with message " + e.msg }

func (e errWithMessage) matchError(e2 error) bool {
	return e2 != nil && e.msg == e2.Error()
}

// CmdExit returns an error that can be passed to Case.Throws to match an
// eval.ExternalCmdExit ignoring the Pid field.
func CmdExit(v eval.ExternalCmdExit) error { return errCmdExit{v} }

// An errorMatcher for an ExternalCmdExit error that ignores the `Pid` member.
// We only match the command name and exit status because at run time we
// cannot know the correct value for `Pid`.
type errCmdExit struct{ v eval.ExternalCmdExit }

func (e errCmdExit) Error() string {
	return e.v.Error()
}

func (e errCmdExit) matchError(gotErr error) bool {
	if gotErr == nil {
		return false
	}
	ge := gotErr.(eval.ExternalCmdExit)
	return e.v.CmdName == ge.CmdName && e.v.WaitStatus == ge.WaitStatus
}

type errOneOf struct{ errs []error }

func OneOfErrors(errs ...error) error { return errOneOf{errs} }

func (e errOneOf) Error() string { return fmt.Sprint("one of", e.errs) }

func (e errOneOf) matchError(gotError error) bool {
	for _, want := range e.errs {
		if matchErr(want, gotError) {
			return true
		}
	}
	return false
}
