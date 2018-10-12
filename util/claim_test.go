package util

import (
	"os"
	"testing"
)

var claimFileTests = []struct {
	pattern      string
	wantFileName string
}{
	{"a*.log", "a9.log"},
	{"*.txt", "1.txt"},
}

func TestClaimFile(t *testing.T) {
	_, cleanup := InTestDir()
	defer cleanup()

	touch("a0.log")
	touch("a1.log")
	touch("a8.log")

	for _, test := range claimFileTests {
		f, err := ClaimFile(".", test.pattern)
		if err != nil {
			t.Errorf("ClaimFile errors: %v", err)
		}
		if f.Name() != test.wantFileName {
			t.Errorf("ClaimFile claims %s, want %s", f.Name(), test.wantFileName)
		}
	}
}

func touch(fname string) {
	f, err := os.Create(fname)
	if err != nil {
		panic(err)
	}
	f.Close()
}
