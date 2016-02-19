package eval

import (
	"errors"
	"os"
	"strings"
)

// Errors
var (
	ErrCanOnlyAssignList          = errors.New("can only assign compatible values")
	ErrPathMustBeString           = errors.New("path must be string")
	ErrPathCannotContainColonZero = errors.New(`path cannot contain colon or \0`)
)

// EnvPathList is a variable whose value is constructed from an environment
// variable by splitting at colons. Changes to it are also propagated to the
// corresponding environment variable. Its elements cannot contain colons or
// \0; attempting to put colon or \0 in its elements will result in an error.
//
// EnvPathList implements both Value and Variable interfaces. It also satisfied
// ListLike.
type EnvPathList struct {
	envName     string
	cachedValue string
	cachedPaths []string
}

var (
	_ Variable = (*EnvPathList)(nil)
	_ Value    = (*EnvPathList)(nil)
	_ ListLike = (*EnvPathList)(nil)
)

func (epl *EnvPathList) Get() Value {
	return epl
}

func (epl *EnvPathList) Set(v Value) {
	elemser, ok := v.(Elemser)
	if !ok {
		throw(ErrCanOnlyAssignList)
	}
	var paths []string
	for v := range elemser.Elems() {
		s, ok := v.(String)
		if !ok {
			throw(ErrPathMustBeString)
		}
		path := string(s)
		if strings.ContainsAny(path, ":\x00") {
			throw(ErrPathCannotContainColonZero)
		}
		paths = append(paths, string(s))
	}
	epl.set(paths)
}

func (epl *EnvPathList) Kind() string {
	return "list"
}

func (epl *EnvPathList) Repr() string {
	epl.sync()
	var b ListReprBuilder
	for _, path := range epl.cachedPaths {
		b.WriteElem(quote(path))
	}
	return b.String()
}

func (epl *EnvPathList) Len() int {
	epl.sync()
	return len(epl.cachedPaths)
}

func (epl *EnvPathList) Elems() <-chan Value {
	ch := make(chan Value)
	go func() {
		epl.sync()
		for _, p := range epl.cachedPaths {
			ch <- String(p)
		}
		close(ch)
	}()
	return ch
}

func (epl *EnvPathList) IndexOne(idx Value) Value {
	epl.sync()
	i := intIndexWithin(idx, len(epl.cachedPaths))
	return String(epl.cachedPaths[i])
}

func (epl *EnvPathList) IndexSet(idx, v Value) {
	epl.sync()
	i := intIndexWithin(idx, len(epl.cachedPaths))
	s, ok := v.(String)
	if !ok {
		throw(ErrPathMustBeString)
	}
	epl.cachedPaths[i] = string(s)
	epl.set(epl.cachedPaths)
}

func (epl *EnvPathList) sync() {
	value := os.Getenv(epl.envName)
	if value == epl.cachedValue {
		return
	}
	epl.cachedValue = value
	epl.cachedPaths = strings.Split(value, ":")
}

func (epl *EnvPathList) get() []string {
	epl.sync()
	return epl.cachedPaths
}

func (epl *EnvPathList) set(paths []string) {
	epl.cachedPaths = paths
	epl.cachedValue = strings.Join(paths, ":")
	err := os.Setenv(epl.envName, epl.cachedValue)
	maybeThrow(err)
}
