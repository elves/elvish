package eval

// Foreign represents a foreign command provider.
type Foreign interface {
	Call(name string, args []Value) Exitus
}

// Editor is the interface through which the Evaler calls the line editor.
type Editor interface {
	Foreign
	Bind(key string, function Value) Exitus
}
