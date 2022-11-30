package errutil

import "strings"

// Multi combines multiple errors into one:
//
//   - If all errors are nil, it returns nil.
//
//   - If there is one non-nil error, it is returned.
//
//   - Otherwise, the return value is an error whose Error methods contain all
//     the messages of all non-nil arguments.
//
// If the input contains any error returned by Multi, such errors are flattened.
// The following two calls return the same value:
//
//	Multi(Multi(err1, err2), Multi(err3, err4))
//	Multi(err1, err2, err3, err4)
func Multi(errs ...error) error {
	var nonNil []error
	for _, err := range errs {
		if err != nil {
			if multi, ok := err.(multiError); ok {
				nonNil = append(nonNil, multi...)
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
		return multiError(nonNil)
	}
}

type multiError []error

func (me multiError) Error() string {
	var sb strings.Builder
	sb.WriteString("multiple errors: ")
	for i, e := range me {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(e.Error())
	}
	return sb.String()
}
