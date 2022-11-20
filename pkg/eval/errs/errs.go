// Package errs declares error types used as exception causes.
package errs

import (
	"fmt"
	"strconv"

	"src.elv.sh/pkg/parse"
)

// OutOfRange encodes an error where a value is out of its valid range.
type OutOfRange struct {
	What      string
	ValidLow  string
	ValidHigh string
	Actual    string
}

// Error implements the error interface.
func (e OutOfRange) Error() string {
	return fmt.Sprintf(
		"out of range: %s must be from %s to %s, but is %s",
		e.What, e.ValidLow, e.ValidHigh, e.Actual)
}

// BadValue encodes an error where the value does not meet a requirement. For
// out-of-range errors, use OutOfRange.
type BadValue struct {
	What   string
	Valid  string
	Actual string
}

// Error implements the error interface.
func (e BadValue) Error() string {
	return fmt.Sprintf(
		"bad value: %v must be %v, but is %v", e.What, e.Valid, e.Actual)
}

// ArityMismatch encodes an error where the expected number of values is out of
// the valid range.
type ArityMismatch struct {
	What      string
	ValidLow  int
	ValidHigh int
	Actual    int
}

func (e ArityMismatch) Error() string {
	switch {
	case e.ValidHigh == e.ValidLow:
		return fmt.Sprintf("arity mismatch: %v must be %v, but is %v",
			e.What, nValues(e.ValidLow), nValues(e.Actual))
	case e.ValidHigh == -1:
		return fmt.Sprintf("arity mismatch: %v must be %v or more values, but is %v",
			e.What, e.ValidLow, nValues(e.Actual))
	default:
		return fmt.Sprintf("arity mismatch: %v must be %v to %v values, but is %v",
			e.What, e.ValidLow, e.ValidHigh, nValues(e.Actual))
	}
}

func nValues(n int) string {
	if n == 1 {
		return "1 value"
	}
	return strconv.Itoa(n) + " values"
}

// SetReadOnlyVar is returned by the Set method of a read-only variable.
type SetReadOnlyVar struct {
	// Name of the read-only variable. This field is initially empty, and
	// populated later when context information is available.
	VarName string
}

// Error implements the error interface.
func (e SetReadOnlyVar) Error() string {
	return fmt.Sprintf(
		"cannot set read-only variable $%s", parse.QuoteVariableName(e.VarName))
}

// ReaderGone is raised by the writer in a pipeline when the reader end has
// terminated. It could be raised directly by builtin commands, or when an
// external command gets terminated by SIGPIPE after Elvish detects the read end
// of the pipe has exited earlier.
type ReaderGone struct {
}

// Error implements the error interface.
func (e ReaderGone) Error() string {
	return "reader gone"
}
