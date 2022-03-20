package eval

// Callable wraps the Call method.
type Callable interface {
	// Call calls the receiver in a Frame with arguments and options.
	Call(fm *Frame, args []any, opts map[string]any) error
}

var (
	// NoArgs is an empty argument list. It can be used as an argument to Call.
	NoArgs = []any{}
	// NoOpts is an empty option map. It can be used as an argument to Call.
	NoOpts = map[string]any{}
)
