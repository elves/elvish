package eval

import (
	"os"

	"src.elv.sh/pkg/eval/vars"
)

// NewPwdVar returns a variable who value is synchronized with the path of the
// current working directory.
func NewPwdVar(ev *Evaler) vars.Var { return pwdVar{ev} }

// pwdVar is a variable whose value always reflects the current working
// directory. Setting it changes the current working directory.
type pwdVar struct {
	ev *Evaler
}

var _ vars.Var = pwdVar{}

var OSGetwd = os.Getwd // to allow mutation by unit tests

// Get returns the current working directory. It returns "/unknown/pwd" when
// it cannot be determined.
func (pwdVar) Get() any {
	pwd, err := OSGetwd()
	if err != nil {
		// This should really use the path separator appropriate for the
		// platform but in practice this hardcoded string works fine. Both
		// because MS Windows supports forward slashes and this will very
		// rarely occur.
		return "/unknown/pwd"
	}
	return pwd
}

// Set changes the current working directory.
func (pwd pwdVar) Set(v any) error {
	path, ok := v.(string)
	if !ok {
		return vars.ErrPathMustBeString
	}
	return pwd.ev.Chdir(path)
}
