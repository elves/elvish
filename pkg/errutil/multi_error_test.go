package errutil

import (
	"errors"
	"testing"
)

var (
	err1 = errors.New("error 1")
	err2 = errors.New("error 2")
	err3 = errors.New("error 3")
)

var errorsTests = []struct {
	e          error
	wantString string
}{
	{Multi(), ""},
	{Multi(errors.New("some error")), "some error"},
	{
		Multi(err1, err2),
		"multiple errors: error 1; error 2",
	},
	{
		Multi(err1, err2, err3),
		"multiple errors: error 1; error 2; error 3",
	},
	{
		Multi(err1, Multi(err2, err3)),
		"multiple errors: error 1; error 2; error 3",
	},
	{
		Multi(Multi(err1, err2), err3),
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
