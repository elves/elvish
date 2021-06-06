package vars

// Note: This doesn't have an associated env_list_tests.go because most of its functionality is
// tested by TestSetEnv_PATH and related tests.

import (
	"errors"
	"os"
	"strings"
	"sync"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/persistent/vector"
)

var (
	pathListSeparator = string(os.PathListSeparator)
	forbiddenInPath   = pathListSeparator + "\x00"
)

// Errors
var (
	ErrPathMustBeString = errors.New("path must be string")
	ErrInvalidPathVal   = errors.New(`path cannot contain (semi)colon or \x00`)
)

// NewEnvListVar returns a variable whose value is a list synchronized with an
// environment variable with the elements joined by os.PathListSeparator.
//
// Elements in the value of the variable must be strings, and cannot contain
// os.PathListSeparator or \0; attempting to put any in its elements will result in
// an error.
func NewEnvListVar(name string) Var {
	return &envListVar{envName: name}
}

type envListVar struct {
	sync.RWMutex
	envName    string
	cacheFor   string
	cacheValue interface{}
}

// Get returns a Value for an EnvPathList.
func (envli *envListVar) Get() interface{} {
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
func (envli *envListVar) Set(v interface{}) error {
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
			errElement = ErrInvalidPathVal
			return false
		}
		paths = append(paths, s)
		return true
	})

	if errElement != nil || errIterate != nil {
		return diag.Errors(errElement, errIterate)
	}

	envli.Lock()
	defer envli.Unlock()
	os.Setenv(envli.envName, strings.Join(paths, pathListSeparator))
	return nil
}
