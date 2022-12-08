package diag

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var (
	err1 = &Error{
		Type:    "foo error",
		Message: "bad 1",
		Context: *contextInParen("[test]", "echo (1)"),
	}
	err2 = &Error{
		Type:    "foo error",
		Message: "bad 2",
		Context: *contextInParen("[test]", "echo (2)"),
	}
)

var cognateErrorsTests = []struct {
	name    string
	errs    []*Error
	wantErr error
}{
	{
		name:    "no error",
		errs:    nil,
		wantErr: nil,
	},
	{
		name:    "one error",
		errs:    []*Error{err1},
		wantErr: err1,
	},
	{
		name:    "multiple errors",
		errs:    []*Error{err1, err2},
		wantErr: cognateErrors{err1, err2},
	},
}

func TestPackAndUnpackCognateErrors(t *testing.T) {
	for _, tc := range cognateErrorsTests {
		t.Run(tc.name, func(t *testing.T) {
			err := PackCognateErrors(tc.errs)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("got packed error %#v, want %#v", err, tc.wantErr)
			}
			unpacked := UnpackCognateErrors(err)
			if !reflect.DeepEqual(unpacked, tc.errs) {
				t.Errorf("Unpacked: (-want +got):\n%s", cmp.Diff(tc.errs, unpacked))
			}
		})
	}
}

func TestUnpackCognateErrors_CalledWithOtherErrorType(t *testing.T) {
	unpacked := UnpackCognateErrors(errors.New("foo"))
	if unpacked != nil {
		t.Errorf("want nil, got %v", unpacked)
	}
}

func TestCognateErrors(t *testing.T) {
	setCulpritMarkers(t, "<", ">")
	setMessageMarkers(t, "{", "}")
	err := PackCognateErrors([]*Error{err1, err2})
	wantError := "multiple foo errors in [test]: 1:6: bad 1; 1:6: bad 2"
	if s := err.Error(); s != wantError {
		t.Errorf(".Error() returns unexpected result (-want +got):\n%s",
			cmp.Diff(wantError, s))
	}
	wantShow := dedent(`
			Multiple foo errors in [test]:
			  {bad 1}
			    1:6: echo <(1)>
			  {bad 2}
			    1:6: echo <(2)>`)
	if show := err.(Shower).Show(""); show != wantShow {
		t.Errorf(".Show(\"\") returns unexpected result (-want +got):\n%s",
			cmp.Diff(wantShow, show))
	}
}
