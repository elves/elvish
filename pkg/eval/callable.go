package eval

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
