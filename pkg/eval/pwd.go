package eval

import (
	"os"

	"github.com/elves/elvish/pkg/eval/vars"
)

// PwdVariable is a variable whose value always reflects the current working
// directory. Setting it changes the current working directory.
type PwdVariable struct {
	ev *Evaler
}

var _ vars.Var = PwdVariable{}

// Get returns the current working directory. It returns /unknown/pwd when it
// cannot be determined.
func (PwdVariable) Get() interface{} {
	pwd, err := os.Getwd()
	// TODO: Deprecate the $pwd variable.
	if err != nil {
		return "/unknown/pwd"
	}
	return pwd
}

// Set changes the current working directory.
func (pwd PwdVariable) Set(v interface{}) error {
	path, ok := v.(string)
	if !ok {
		return ErrPathMustBeString
	}
	return pwd.ev.Chdir(path)
}
