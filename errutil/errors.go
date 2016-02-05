package errutil

import "bytes"

type Errors struct {
	Errors []error
}

func (es *Errors) Error() string {
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

func (es *Errors) Append(e error) {
	es.Errors = append(es.Errors, e)
}
