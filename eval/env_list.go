package eval

import (
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/xiaq/persistent/vector"
)

// Errors
var (
	ErrCanOnlyAssignList          = errors.New("can only assign compatible values")
	ErrPathMustBeString           = errors.New("path must be string")
	ErrPathCannotContainColonZero = errors.New(`path cannot contain colon or \0`)
)

// EnvList is a variable whose value is constructed from an environment
// variable by splitting at colons. Changes to it are also propagated to the
// corresponding environment variable. Its elements cannot contain colons or
// \0; attempting to put colon or \0 in its elements will result in an error.
type EnvList struct {
	sync.RWMutex
	envName    string
	cacheFor   string
	cacheValue Value
}

var (
	_ Variable = (*EnvList)(nil)
)

// Get returns a Value for an EnvPathList.
func (envli *EnvList) Get() Value {
	envli.Lock()
	defer envli.Unlock()

	value := os.Getenv(envli.envName)
	if value == envli.cacheFor {
		return envli.cacheValue
	}
	envli.cacheFor = value
	v := vector.Empty
	for _, path := range strings.Split(value, ":") {
		v = v.Cons(String(path))
	}
	envli.cacheValue = List{v}
	return envli.cacheValue
}

// Set sets an EnvPathList. The underlying environment variable is set.
func (envli *EnvList) Set(v Value) {
	iterator, ok := v.(Iterable)
	if !ok {
		throw(ErrCanOnlyAssignList)
	}
	var paths []string
	iterator.Iterate(func(v Value) bool {
		s, ok := v.(String)
		if !ok {
			throw(ErrPathMustBeString)
		}
		path := string(s)
		if strings.ContainsAny(path, ":\x00") {
			throw(ErrPathCannotContainColonZero)
		}
		paths = append(paths, string(s))
		return true
	})

	envli.Lock()
	defer envli.Unlock()
	os.Setenv(envli.envName, strings.Join(paths, ":"))
}
