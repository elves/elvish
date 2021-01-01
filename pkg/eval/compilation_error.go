package eval

import "github.com/elves/elvish/pkg/diag"

const compilationErrorType = "compilation error"

// NewCompilationError creates a new compilation error.
func NewCompilationError(message string, context diag.Context) error {
	return &diag.Error{
		Type: compilationErrorType, Message: message, Context: context}
}

// GetCompilationError returns a *diag.Error if the given value is a compilation
// error. Otherwise it returns nil.
func GetCompilationError(e interface{}) *diag.Error {
	if e, ok := e.(*diag.Error); ok && e.Type == compilationErrorType {
		return e
	}
	return nil
}
