package eval

import (
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/util"
	"github.com/xiaq/persistent/vector"
)

var (
	pathListSeparator = string(os.PathListSeparator)
	forbiddenInPath   = pathListSeparator + "\x00"
)

// Errors
var (
	ErrCanOnlyAssignList          = errors.New("can only assign compatible values")
	ErrPathMustBeString           = errors.New("path must be string")
	ErrPathCannotContainColonZero = errors.New(`path cannot contain colon or \0`)
)

// EnvList is a variable whose value is constructed from an environment variable
// by splitting at pathListSeparator. Changes to it are also propagated to the
// corresponding environment variable. Its elements cannot contain
// pathListSeparator or \0; attempting to put any in its elements will result in
// an error.
type EnvList struct {
	sync.RWMutex
	envName    string
	cacheFor   string
	cacheValue interface{}
}

var (
	_ vars.Var = (*EnvList)(nil)
)

// Get returns a Value for an EnvPathList.
func (envli *EnvList) Get() interface{} {
	envli.Lock()
	defer envli.Unlock()

	value := os.Getenv(envli.envName)
	if value == envli.cacheFor {
		return envli.cacheValue
	}
	envli.cacheFor = value
	v := vector.Empty
	for _, path := range strings.Split(value, pathListSeparator) {
		v = v.Cons(path)
	}
	envli.cacheValue = v
	return envli.cacheValue
}

// Set sets an EnvPathList. The underlying environment variable is set.
func (envli *EnvList) Set(v interface{}) error {
	var (
		paths      []string
		errElement error
	)
	errIterate := vals.Iterate(v, func(v interface{}) bool {
		s, ok := v.(string)
		if !ok {
			errElement = ErrPathMustBeString
			return false
		}
		path := s
		if strings.ContainsAny(path, forbiddenInPath) {
			errElement = ErrPathCannotContainColonZero
			return false
		}
		paths = append(paths, s)
		return true
	})

	if errElement != nil || errIterate != nil {
		return util.Errors(errElement, errIterate)
	}

	envli.Lock()
	defer envli.Unlock()
	os.Setenv(envli.envName, strings.Join(paths, pathListSeparator))
	return nil
}
