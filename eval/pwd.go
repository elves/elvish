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

func (PwdVariable) Set(v Value) {
	path, ok := v.(String)
	if !ok {
		throw(ErrPathMustBeString)
	}
	err := os.Chdir(string(path))
	maybeThrow(err)
}
