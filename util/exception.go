package util

// This file provides an exception-like mechanism.

// An exception wraps an error.
type exception struct {
	err error
}

// Panic panics with err wrapped properly so that it can be catched by Recover.
func Panic(err error) {
	panic(exception{err})
}

// Recover tries to catch an error thrown by Panic and stop the panic. If the
// panic is not caused by Panic, the panic is not stopped.
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
