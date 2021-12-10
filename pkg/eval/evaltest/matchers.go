package evaltest

import (
	"fmt"
	"math"
	"reflect"
	"regexp"

	"src.elv.sh/pkg/eval"
)

// ApproximatelyThreshold defines the threshold for matching float64 values when
// using Approximately.
const ApproximatelyThreshold = 1e-15

// Approximately can be passed to Case.Puts to match a float64 within the
// threshold defined by ApproximatelyThreshold.
type Approximately struct{ F float64 }

func matchFloat64(a, b, threshold float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	if math.IsInf(a, 0) && math.IsInf(b, 0) &&
		math.Signbit(a) == math.Signbit(b) {
		return true
	}
	return math.Abs(a-b) <= threshold
}

// MatchingRegexp can be passed to Case.Puts to match a any string that matches
// a regexp pattern. If the pattern is not a valid regexp, the test will panic.
type MatchingRegexp struct{ Pattern string }

func matchRegexp(p, s string) bool {
	matched, err := regexp.MatchString(p, s)
	if err != nil {
		panic(err)
	}
	return matched
}

type errorMatcher interface{ matchError(error) bool }

// An errorMatcher for any error.
type anyError struct{}

func (anyError) Error() string { return "any error" }

func (anyError) matchError(e error) bool { return e != nil }

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
		texts = append(texts, ctx.Source[ctx.From:ctx.To])
		tb = tb.Next
	}
	return texts
}

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
