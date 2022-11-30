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
		Context: *parseContext("echo [1]", "[", "]", true),
	}
	err2 = &Error{
		Type:    "foo error",
		Message: "bad 2",
		Context: *parseContext("echo [2]", "[", "]", true),
	}
)

var cognateErrorsTests = []struct {
	name      string
	errs      []*Error
	wantError string
	wantShow  string
}{
	{
		name:      "no error",
		errs:      nil,
		wantError: "",
		wantShow:  "",
	},
	{
		name:      "one error",
		errs:      []*Error{err1},
		wantError: err1.Error(),
		wantShow:  err1.Show(""),
	},
	{
		name:      "multiple errors",
		errs:      []*Error{err1, err2},
		wantError: "multiple foo errors in [test]: 5-8: bad 1; 5-8: bad 2",
		wantShow: lines(
			"Multiple foo errors in [test]:",
			"  \x1b[31;1mbad 1\x1b[m",
			"    line 1: echo \x1b[1;4m[1]\x1b[m",
			"  \x1b[31;1mbad 2\x1b[m",
			"    line 1: echo \x1b[1;4m[2]\x1b[m"),
	},
}

func TestCognateErrors(t *testing.T) {
	for _, tc := range cognateErrorsTests {
		t.Run(tc.name, func(t *testing.T) {
			err := PackCognateErrors(tc.errs)
			if err == nil {
				if tc.wantError != "" || tc.wantShow != "" {
					t.Errorf("Want non-nil error, got nil")
				}
			} else {
				if got := err.Error(); got != tc.wantError {
					t.Errorf("Error() (-want +got):\n%s", cmp.Diff(tc.wantError, got))
				}
				if got := err.(Shower).Show(""); got != tc.wantShow {
					t.Errorf("Show() (-want +got):\n%s", cmp.Diff(tc.wantShow, got))
				}
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
