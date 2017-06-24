package eval

import (
	"os"

	"github.com/elves/elvish/daemon/api"
)

// PwdVariable is a variable whose value always reflects the current working
// directory. Setting it changes the current working directory.
type PwdVariable struct {
	daemon *api.Client
}

var _ Variable = PwdVariable{}

func (PwdVariable) Get() Value {
	pwd, err := os.Getwd()
	maybeThrow(err)
	return String(pwd)
}

func (pwd PwdVariable) Set(v Value) {
	path, ok := v.(String)
	if !ok {
		throw(ErrPathMustBeString)
	}
	err := Chdir(string(path), pwd.daemon)
	maybeThrow(err)
}
