// +build wtf

package eval_test

import (
	"errors"
	"path/filepath"
	"testing"

	. "github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/testutil"

	. "github.com/elves/elvish/pkg/eval/evaltest"
)

func TestBuiltinPwd(t *testing.T) {
	tmpHome, cleanup := testutil.InTempHome()
	defer cleanup()

	testutil.MustMkdirAll("dir1")
	testutil.MustMkdirAll("dir2")
	dir1 := filepath.Join(tmpHome, "dir1")
	dir2 := filepath.Join(tmpHome, "dir2")

	Test(t,
		That(`pwd=dir1 put $pwd; put $pwd`).Puts(dir1, tmpHome),
		That(`cd ~/dir2; pwd=~ put $pwd; put $pwd`).Puts(tmpHome, dir2),
		That(`cd ~/dir2; pwd=~/dir1 put $pwd; put $pwd`).Puts(dir1, dir2),
		That(`cd ~/dir1; pwd=../dir2 put $pwd; put $pwd`).Puts(dir2, dir1),
		That(`pwd=(float64 1) put $pwd`).Throws(ErrPathMustBeString, "pwd=(float64 1)"),
	)
}

// Verify the behavior when the CWD cannot be determined.
func TestBuiltinPwd_GetwdError(t *testing.T) {
	origGetwd := getwd
	getwd = mockGetwdWithError
	defer func() { getwd = origGetwd }()

	Test(t,
		That(`put $pwd`).Puts("/unknown/pwd"),
	)
}

func mockGetwdWithError() (string, error) {
	return "", errors.New("cwd unknown")
}
