package util

// This file provides an exception-like mechanism.

// An exception wraps an error.
type exception struct {
	err error
}

func Panic(err error) {
	panic(exception{err})
}

func Recover(perr *error) {
	r := recover()
	if r == nil {
		return
	}
	if exc, ok := r.(exception); ok {
		*perr = exc.err
	} else {
		panic(r)
	}
}
