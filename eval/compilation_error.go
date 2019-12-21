package eval

import "github.com/elves/elvish/diag"

const compilationErrorType = "compilation error"

// NewCompilationError creates a new compilation error.
func NewCompilationError(message string, context diag.Context) error {
	return &diag.Error{
		Type: compilationErrorType, Message: message, Context: context}
}

// GetCompilationError returns a *diag.Error and true if the given error is a
// compilation error. Otherwise it returns nil and false.
func GetCompilationError(e error) (*diag.Error, bool) {
	if e, ok := e.(*diag.Error); ok && e.Type == compilationErrorType {
		return e, true
	}
	return nil, false
}
