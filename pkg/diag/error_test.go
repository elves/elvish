package diag

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type fooErrorTag struct{}

func (fooErrorTag) ErrorTag() string { return "foo error" }

func TestError(t *testing.T) {
	setContextBodyMarkers(t, "<", ">")
	setMessageMarkers(t, "{", "}")

	err := &Error[fooErrorTag]{
		Message: "bad list",
		Context: *contextInParen("[test]", "echo (x)"),
	}

	wantErrorString := "foo error: [test]:1:6-8: bad list"
	if gotErrorString := err.Error(); gotErrorString != wantErrorString {
		t.Errorf("Error() -> %q, want %q", gotErrorString, wantErrorString)
	}

	wantRanging := Ranging{From: 5, To: 8}
	if gotRanging := err.Range(); gotRanging != wantRanging {
		t.Errorf("Range() -> %v, want %v", gotRanging, wantRanging)
	}

	// Title() is used for Show
	wantShow := dedent(`
		Foo error: {bad list}
		  [test]:1:6-8: echo <(x)>`)
	if gotShow := err.Show(""); gotShow != wantShow {
		t.Errorf("Show() -> %q, want %q", gotShow, wantShow)
	}
}

var (
	err1 = &Error[fooErrorTag]{
		Message: "bad 1",
		Context: *contextInParen("a.elv", "echo (1)"),
	}
	err2 = &Error[fooErrorTag]{
		Message: "bad 2",
		Context: *contextInParen("b.elv", "echo (2\n0)"),
	}
)

var multiErrorsTests = []struct {
	name    string
	errs    []*Error[fooErrorTag]
	wantErr error
}{
	{
		name:    "no error",
		errs:    nil,
		wantErr: nil,
	},
	{
		name:    "one error",
		errs:    []*Error[fooErrorTag]{err1},
		wantErr: err1,
	},
	{
		name:    "multiple errors",
		errs:    []*Error[fooErrorTag]{err1, err2},
		wantErr: multiError[fooErrorTag]{err1, err2},
	},
}

func TestPackAndUnpackErrors(t *testing.T) {
	for _, tc := range multiErrorsTests {
		t.Run(tc.name, func(t *testing.T) {
			err := PackErrors(tc.errs)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("got packed error %#v, want %#v", err, tc.wantErr)
			}
			unpacked := UnpackErrors[fooErrorTag](err)
			if !reflect.DeepEqual(unpacked, tc.errs) {
				t.Errorf("Unpacked: (-want +got):\n%s", cmp.Diff(tc.errs, unpacked))
			}
		})
	}
}

func TestUnpackErrors_CalledWithOtherErrorType(t *testing.T) {
	unpacked := UnpackErrors[fooErrorTag](errors.New("foo"))
	if unpacked != nil {
		t.Errorf("want nil, got %v", unpacked)
	}
}

func TestMultiError_ErrorAndShow(t *testing.T) {
	setContextBodyMarkers(t, "<", ">")
	setMessageMarkers(t, "{", "}")
	err := PackErrors([]*Error[fooErrorTag]{err1, err2})
	wantError := "multiple foo errors: a.elv:1:6-8: bad 1; b.elv:1:6-2:2: bad 2"
	if s := err.Error(); s != wantError {
		t.Errorf(".Error() returns unexpected result (-want +got):\n%s",
			cmp.Diff(wantError, s))
	}
	wantShow := dedent(`
			Multiple foo errors:
			  {bad 1}
			    a.elv:1:6-8: echo <(1)>
			  {bad 2}
			    b.elv:1:6-2:2:
			      echo <(2>
			      <0)>`)
	if show := err.(Shower).Show(""); show != wantShow {
		t.Errorf(".Show(\"\") returns unexpected result (-want +got):\n%s",
			cmp.Diff(wantShow, show))
	}
}
