package eval

// Editor is the interface through which the Evaler calls the line editor.
type Editor interface {
	Bind(key string, function Value) error
}
