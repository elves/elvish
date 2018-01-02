package eval

import (
	"os"

	"github.com/elves/elvish/eval/types"
)

// PwdVariable is a variable whose value always reflects the current working
// directory. Setting it changes the current working directory.
type PwdVariable struct {
	store AddDirer
}

var _ Variable = PwdVariable{}

func (PwdVariable) Get() types.Value {
	pwd, err := os.Getwd()
	maybeThrow(err)
	return types.String(pwd)
}

func (pwd PwdVariable) Set(v types.Value) {
	path, ok := v.(types.String)
	if !ok {
		throw(ErrPathMustBeString)
	}
	err := Chdir(string(path), pwd.store)
	maybeThrow(err)
}
