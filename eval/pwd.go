package eval

import (
	"os"

	"github.com/elves/elvish/eval/vars"
)

// PwdVariable is a variable whose value always reflects the current working
// directory. Setting it changes the current working directory.
type PwdVariable struct {
	ev *Evaler
}

var _ vars.Var = PwdVariable{}

func (PwdVariable) Get() interface{} {
	pwd, err := os.Getwd()
	maybeThrow(err)
	return pwd
}

func (pwd PwdVariable) Set(v interface{}) error {
	path, ok := v.(string)
	if !ok {
		return ErrPathMustBeString
	}
	return pwd.ev.Chdir(path)
}
