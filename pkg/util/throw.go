package util

// Thrown wraps an error that was raised by Throw, so that it can be recognized
// by Catch.
type Thrown struct {
	Wrapped error
}

func (t Thrown) Error() string {
	return "thrown: " + t.Wrapped.Error()
}

// Throw panics with err wrapped properly so that it can be catched by Catch.
func Throw(err error) {
	panic(Thrown{err})
}

// Catch tries to catch an error thrown by Throw and stop the panic. If the
// panic is not caused by Throw, the panic is not stopped. It should be called
// directly from defer.
func Catch(perr *error) {
	r := recover()
	if r == nil {
		return
	}
	if exc, ok := r.(Thrown); ok {
		*perr = exc.Wrapped
	} else {
		panic(r)
	}
}

// PCall calls a function and catches anything Thrown'n and returns it. It does
// not protect against panics not using Throw, nor can it distinguish between
// nothing thrown and Throw(nil).
func PCall(f func()) (e error) {
	defer Catch(&e)
	f()
	// If we reach here, f didn't throw anything.
	return nil
}

// Throws returns whether calling f throws out a certain error (using Throw). It
// is useful for testing.
func Throws(f func(), e error) bool {
	return PCall(f) == e
}

// ThrowsAny returns whether calling f throws out anything that is not nil. It
// is useful for testing.
func ThrowsAny(f func()) bool {
	return PCall(f) != nil
}

// DoesntThrow returns whether calling f does not throw anything. It is useful
// for testing.
func DoesntThrow(f func()) bool {
	return PCall(f) == nil
}
