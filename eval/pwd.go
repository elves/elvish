package eval

import (
	"os"

	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vartypes"
)

// PwdVariable is a variable whose value always reflects the current working
// directory. Setting it changes the current working directory.
type PwdVariable struct {
	store AddDirer
}

var _ vartypes.Variable = PwdVariable{}

func (PwdVariable) Get() types.Value {
	pwd, err := os.Getwd()
	maybeThrow(err)
	return pwd
}

func (pwd PwdVariable) Set(v types.Value) error {
	path, ok := v.(string)
	if !ok {
		return ErrPathMustBeString
	}
	return Chdir(path, pwd.store)
}
