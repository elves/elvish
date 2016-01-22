package errutil

// type marker for exceptions
type exception struct {
	error
}

// Throw panics with err wrapped properly so that it can be catched by Catch.
func Throw(err error) {
	panic(exception{err})
}

// Catch tries to catch an error thrown by Throw and stop the panic. If the
// panic is not caused by Throw, the panic is not stopped.
func Catch(perr *error) {
	r := recover()
	if r == nil {
		return
	}
	if exc, ok := r.(exception); ok {
		*perr = exc.error
	} else {
		panic(r)
	}
}
