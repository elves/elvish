package vars

// Note: This doesn't have an associated env_list_tests.go because most of its functionality is
// tested by TestSetEnv_PATH and related tests.

import (
	"errors"
	"os"
	"strings"
	"sync"

	"src.elv.sh/pkg/errutil"
	"src.elv.sh/pkg/eval/vals"
)

var (
	pathListSeparator = string(os.PathListSeparator)
	forbiddenInPath   = pathListSeparator + "\x00"
)

// Errors
var (
	ErrPathMustBeString          = errors.New("path must be string")
	ErrPathContainsForbiddenChar = errors.New("path cannot contain NUL byte, colon on Unix or semicolon on Windows")
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
	cacheValue any
}

// Get returns a Value for an EnvPathList.
func (envli *envListVar) Get() any {
	envli.Lock()
	defer envli.Unlock()

	value := os.Getenv(envli.envName)
	if value == envli.cacheFor {
		return envli.cacheValue
	}
	envli.cacheFor = value
	v := vals.EmptyList
	for _, path := range strings.Split(value, pathListSeparator) {
		v = v.Conj(path)
	}
	envli.cacheValue = v
	return envli.cacheValue
}

// Set sets an EnvPathList. The underlying environment variable is set.
func (envli *envListVar) Set(v any) error {
	var (
		paths      []string
		errElement error
	)
	errIterate := vals.Iterate(v, func(v any) bool {
		s, ok := v.(string)
		if !ok {
			errElement = ErrPathMustBeString
			return false
		}
		path := s
		if strings.ContainsAny(path, forbiddenInPath) {
			errElement = ErrPathContainsForbiddenChar
			return false
		}
		paths = append(paths, s)
		return true
	})

	if errElement != nil || errIterate != nil {
		return errutil.Multi(errElement, errIterate)
	}

	envli.Lock()
	defer envli.Unlock()
	os.Setenv(envli.envName, strings.Join(paths, pathListSeparator))
	return nil
}
