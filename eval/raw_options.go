package eval

import "testing"

type opts struct {
	FooBar string
	POSIX  bool `name:"posix"`
	Min    int
}

var scanOptionsTests = []struct {
	rawOpts  RawOptions
	preScan  opts
	postScan opts
}{
	{RawOptions{"foo-bar": "lorem ipsum"},
		opts{}, opts{FooBar: "lorem ipsum"}},
	{RawOptions{"posix": true},
		opts{}, opts{POSIX: true}},
}

func TestScanOptions(t *testing.T) {
	for _, test := range scanOptionsTests {
		opts := test.preScan
		scanOptions(test.rawOpts, &opts)
		if opts != test.postScan {
			t.Errorf("Scan %v => %v, want %v", test.rawOpts, opts, test.postScan)
		}
	}
}
