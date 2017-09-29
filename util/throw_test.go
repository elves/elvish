package util

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

func TestThrowAndCatch(t *testing.T) {
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
	_ = recoverPanic(f)
	if err != nil {
		t.Errorf("Catch recovered panic not caused via Throw")
	}

	// Catch should do nothing when there is no panic
	err = nil
	f = func() {
		defer Catch(&err)
	}
	f()
	if err != nil {
		t.Errorf("Catch recovered something when there is no panic")
	}
}

// errToThrow is the error to throw in test cases.
var errToThrow = errors.New("error to throw")

func TestPCall(t *testing.T) {
	// PCall catches throws
	if PCall(func() { Throw(errToThrow) }) != errToThrow {
		t.Errorf("PCall does not catch throws")
	}
	// PCall returns nil when nothing has been thrown
	if PCall(func() {}) != nil {
		t.Errorf("PCall returns non-nil when nothing has been thrown")
	}
	// PCall returns nil when nil has been thrown
	if PCall(func() { Throw(nil) }) != nil {
		t.Errorf("PCall returns non-nil when nil has been thrown")
	}
}

func TestThrows(t *testing.T) {
	if Throws(func() { Throw(errToThrow) }, errToThrow) != true {
		t.Errorf("Throws returns false when function throws wanted error")
	}
	if Throws(func() { Throw(errToThrow) }, errors.New("")) != false {
		t.Errorf("Throws returns true when function throws unwanted error")
	}
	if Throws(func() {}, errToThrow) != false {
		t.Errorf("Throws returns true when function does not throw")
	}
}

func TestThrowsAny(t *testing.T) {
	if Throws(func() { Throw(errToThrow) }, errToThrow) != true {
		t.Errorf("ThrowsAny returns false when function throws non-nil")
	}
	if Throws(func() {}, errToThrow) != false {
		t.Errorf("ThrowsAny returns true when function does not throw")
	}
}

func TestDoesnotThrow(t *testing.T) {
	if DoesntThrow(func() { Throw(errToThrow) }) != false {
		t.Errorf("DoesntThrow returns true when function throws")
	}
	if DoesntThrow(func() {}) != true {
		t.Errorf("DoesntThrow returns false when function doesn't throw")
	}
}
