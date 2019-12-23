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
// returns nil. If there is one non-nil error, it is returned. Otherwise the
// return value is a MultiError containing all the non-nil arguments. Arguments
// of the type MultiError are flattened.
func Errors(errs ...error) error {
	var nonNil []error
	for _, err := range errs {
		if err != nil {
			if multi, ok := err.(MultiError); ok {
				nonNil = append(nonNil, multi.Errors...)
			} else {
				nonNil = append(nonNil, err)
			}
		}
	}
	switch len(nonNil) {
	case 0:
		return nil
	case 1:
		return nonNil[0]
	default:
		return MultiError{nonNil}
	}
}
