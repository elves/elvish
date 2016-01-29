package eval

// Foreign represents a foreign command provider.
type Foreign interface {
	Call(name string, args []Value) Exitus
}
