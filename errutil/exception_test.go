package errutil

import (
	"errors"
	"testing"
)

func recoverPanic(f func()) (recovered interface{}) {
	defer func() {
		recovered = recover()
	}()
	f()
	return nil
}

func TestException(t *testing.T) {
	tothrow := errors.New("an error to throw")
	// Throw should cause a panic
	f := func() {
		Throw(tothrow)
	}
	if recoverPanic(f) == nil {
		t.Errorf("Throw did not cause a panic")
	}

	// Catch should catch what was thrown
	caught := func() (err error) {
		defer Catch(&err)
		Throw(tothrow)
		return nil
	}()
	if caught != tothrow {
		t.Errorf("thrown %v, but caught %v", tothrow, caught)
	}

	// Catch should not recover panics not caused by Throw
	var err error
	f = func() {
		defer Catch(&err)
		panic(errors.New("233"))
	}
	recoverPanic(f)
	if err != nil {
		t.Errorf("Catch recovered panic not caused via Throw")
	}
}
