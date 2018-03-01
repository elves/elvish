package eval

import (
	"fmt"

	"github.com/elves/elvish/eval/vals"
)

// Callable wraps the Call method.
type Callable interface {
	// Call calls the receiver in a Frame with arguments and options.
	Call(fm *Frame, args []interface{}, opts map[string]interface{}) error
}

var (
	// NoArgs is an empty argument list. It can be used as an argument to Call.
	NoArgs = []interface{}{}
	// NoOpts is an empty option map. It can be used as an argument to Call.
	NoOpts = map[string]interface{}{}
)

// FromJSONInterface converts a interface{} that results from json.Unmarshal to
// a Value.
func FromJSONInterface(v interface{}) interface{} {
	if v == nil {
		// TODO Use a more appropriate type
		return ""
	}
	switch v := v.(type) {
	case bool, string:
		return v
	case float64:
		// TODO Use a numeric type for float64
		return fmt.Sprint(v)
	case []interface{}:
		vs := make([]interface{}, len(v))
		for i, v := range v {
			vs[i] = FromJSONInterface(v)
		}
		return vals.MakeList(vs...)
	case map[string]interface{}:
		m := vals.EmptyMap
		for key, val := range v {
			m = m.Assoc(key, FromJSONInterface(val))
		}
		return m
	default:
		throw(fmt.Errorf("unexpected json type: %T", v))
		return nil // not reached
	}
}
