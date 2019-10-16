package shell

import (
	"encoding/json"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/parse"
)

// ErrorInJSON is the data structure for self-defined error
type ErrorInJSON struct {
	FileName string `json:"fileName"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
	Message  string `json:"message"`
}

// SimplifiedErrorInJSON is the data structure for errors which contain limited info
type SimplifiedErrorInJSON struct {
	Message string `json:"message"`
}

// ErrorToJSON converts the error into JSON format
func ErrorToJSON(err error) string {
	var e interface{}
	switch err := err.(type) {
	case *eval.CompilationError:
		e = []interface{}{ErrorInJSON{err.Context.Name, err.Context.Begin, err.Context.End, err.Message}}
	case *parse.MultiError:
		e = processMultiError(err)
	default:
		e = []interface{}{SimplifiedErrorInJSON{err.Error()}}
	}
	jsonError, er := json.Marshal(e)
	if er != nil {
		return `[{"message":"Unable to convert the errors to JSON format"}]`
	}
	return string(jsonError)
}

func processMultiError(e *parse.MultiError) []ErrorInJSON {
	var errArr []ErrorInJSON
	for _, v := range e.Entries {
		errArr = append(errArr, ErrorInJSON{v.Context.Name, v.Context.Begin, v.Context.End, v.Message})
	}
	return errArr
}
