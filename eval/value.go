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

// Converts a interface{} that results from json.Unmarshal to an Elvish value.
func fromJSONInterface(v interface{}) (interface{}, error) {
	if v == nil {
		// TODO Use a more appropriate type
		return "", nil
	}
	switch v := v.(type) {
	case bool, string:
		return v, nil
	case float64:
		// TODO Use a numeric type for float64
		return fmt.Sprint(v), nil
	case []interface{}:
		vec := vals.EmptyList
		for _, elem := range v {
			converted, err := fromJSONInterface(elem)
			if err != nil {
				return nil, err
			}
			vec = vec.Cons(converted)
		}
		return vec, nil
	case map[string]interface{}:
		m := vals.EmptyMap
		for key, val := range v {
			convertedVal, err := fromJSONInterface(val)
			if err != nil {
				return nil, err
			}
			m = m.Assoc(key, convertedVal)
		}
		return m, nil
	default:
		return nil, fmt.Errorf("unexpected json type: %T", v)
	}
}
