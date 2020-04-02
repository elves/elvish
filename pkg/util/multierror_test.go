package util

import (
	"errors"
	"testing"
)

var errorsTests = []struct {
	e          error
	wantString string
}{
	{Errors(), ""},
	{MultiError{}, "no error"},
	{Errors(errors.New("some error")), "some error"},
	{
		Errors(errors.New("error 1"), errors.New("error 2")),
		"multiple errors: error 1; error 2",
	},
	{
		Errors(errors.New("error 1"), errors.New("error 2"), errors.New("error 3")),
		"multiple errors: error 1; error 2; error 3",
	},
	{
		Errors(errors.New("error 1"),
			Errors(errors.New("error 2"), errors.New("error 3"))),
		"multiple errors: error 1; error 2; error 3",
	},
	{
		Errors(Errors(errors.New("error 1"), errors.New("error 2")),
			errors.New("error 3")),
		"multiple errors: error 1; error 2; error 3",
	},
}

func TestErrors(t *testing.T) {
	for _, test := range errorsTests {
		if test.e == nil {
			if test.wantString != "" {
				t.Errorf("got nil, want %q", test.wantString)
			}
		} else {
			gotString := test.e.Error()
			if gotString != test.wantString {
				t.Errorf("got %q, want %q", gotString, test.wantString)
			}
		}
	}
}
