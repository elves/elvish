package testutil

import (
	"io/fs"
	"os"
)

// ChmodOrSkip runs [os.Chmod], but skips the test if file's mode is not exactly
// mode or if there is any error.
func ChmodOrSkip(s Skipper, name string, mode fs.FileMode) {
	err := os.Chmod(name, mode)
	if err != nil {
		s.Skipf("chmod: %v", err)
	}
	fi, err := os.Stat(name)
	if err != nil {
		s.Skipf("stat: %v", err)
	}
	if fi.Mode() != mode {
		s.Skipf("file mode %O is not %O", fi.Mode(), mode)
	}
}
