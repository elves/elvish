package util

import "bytes"

// MultiError pack multiple errors into one error.
type MultiError struct {
	Errors []error
}

func (es MultiError) Error() string {
	switch len(es.Errors) {
	case 0:
		return "no error"
	case 1:
		return es.Errors[0].Error()
	default:
		var buf bytes.Buffer
		buf.WriteString("multiple errors: ")
		for i, e := range es.Errors {
			if i > 0 {
				buf.WriteString("; ")
			}
			buf.WriteString(e.Error())
		}
		return buf.String()
	}
}

// Errors concatenate multiple errors into one. If all errors are nil, it
// returns nil, otherwise the return value is a MultiError containing all the
// non-nil arguments.
func Errors(errs ...error) error {
	var nonNil []error
	for _, err := range errs {
		if err != nil {
			nonNil = append(nonNil, err)
		}
	}
	if len(nonNil) == 0 {
		return nil
	}
	return MultiError{nonNil}
}
