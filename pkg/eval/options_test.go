package eval

import (
	"testing"
)

type opts struct {
	FooBar string
	POSIX  bool `name:"posix"`
	Min    int
	ignore bool // this should be ignored since it isn't exported
}

var scanOptionsTests = []struct {
	rawOpts  RawOptions
	preScan  opts
	postScan opts
	err      error
}{
	{RawOptions{"foo-bar": "lorem ipsum"},
		opts{}, opts{FooBar: "lorem ipsum"}, nil},
	{RawOptions{"posix": true},
		opts{}, opts{POSIX: true}, nil},
	// Since "ignore" is not exported it will result in an error when used.
	{RawOptions{"ignore": true},
		opts{}, opts{ignore: false}, UnknownOption{"ignore"}},
}

func TestScanOptions(t *testing.T) {
	// scanOptions requires a pointer to struct.
	err := scanOptions(RawOptions{}, opts{})
	if err == nil {
		t.Errorf("Scan should have reported invalid options arg error")
	}

	for _, test := range scanOptionsTests {
		opts := test.preScan
		err := scanOptions(test.rawOpts, &opts)

		if ((err == nil) != (test.err == nil)) ||
			(err != nil && test.err != nil && err.Error() != test.err.Error()) {
			t.Errorf("Scan error mismatch %v: want %q, got %q", test.rawOpts, test.err, err)
		}
		if opts != test.postScan {
			t.Errorf("Scan %v => %v, want %v", test.rawOpts, opts, test.postScan)
		}
	}
}
