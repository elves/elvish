package eval

import "testing"

type opts struct {
	FooBar string
	POSIX  bool `name:"posix"`
	Min    int
}

var scanTests = []struct {
	rawOpts  RawOptions
	preScan  opts
	postScan opts
}{
	{RawOptions{"foo-bar": "lorem ipsum"},
		opts{}, opts{FooBar: "lorem ipsum"}},
	{RawOptions{"posix": true},
		opts{}, opts{POSIX: true}},
}

func TestScan(t *testing.T) {
	for _, test := range scanTests {
		opts := test.preScan
		test.rawOpts.Scan(&opts)
		if opts != test.postScan {
			t.Errorf("Scan %v => %v, want %v", test.rawOpts, opts, test.postScan)
		}
	}
}
