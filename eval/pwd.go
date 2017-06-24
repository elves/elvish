package eval

import "os"

// PwdVariable is a variable whose value always reflects the current working
// directory. Setting it changes the current working directory.
type PwdVariable struct{}

var _ Variable = PwdVariable{}

func (PwdVariable) Get() Value {
	pwd, err := os.Getwd()
	maybeThrow(err)
	return String(pwd)
}

// TODO(xiaq): Setting $pwd should also record the new directory in the history.
func (PwdVariable) Set(v Value) {
	path, ok := v.(String)
	if !ok {
		throw(ErrPathMustBeString)
	}
	err := Chdir(string(path), nil)
	maybeThrow(err)
}
