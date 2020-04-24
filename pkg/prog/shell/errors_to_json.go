package shell

import (
	"encoding/json"

	"github.com/elves/elvish/pkg/diag"
	"github.com/elves/elvish/pkg/parse"
)

// An auxiliary struct for converting errors with diagnostics information to JSON.
type errorInJSON struct {
	FileName string `json:"fileName"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
	Message  string `json:"message"`
}

// An auxiliary struct for converting errors with only a message to JSON.
type simpleErrorInJSON struct {
	Message string `json:"message"`
}

// Converts the error into JSON.
func errorToJSON(err error) []byte {
	var e interface{}
	switch err := err.(type) {
	case *diag.Error:
		e = []interface{}{
			errorInJSON{err.Context.Name, err.Context.From, err.Context.To, err.Message},
		}
	case *parse.MultiError:
		var errArr []errorInJSON
		for _, v := range err.Entries {
			errArr = append(errArr,
				errorInJSON{v.Context.Name, v.Context.From, v.Context.To, v.Message})
		}
		e = errArr
	default:
		e = []interface{}{simpleErrorInJSON{err.Error()}}
	}
	jsonError, errMarshal := json.Marshal(e)
	if errMarshal != nil {
		return []byte(`[{"message":"Unable to convert the errors to JSON"}]`)
	}
	return jsonError
}
