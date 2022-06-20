package eval_test

import (
	"errors"
	"path/filepath"
	"runtime"
	"testing"

	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/testutil"

	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vars"
)

func TestBuiltinPwd(t *testing.T) {
	tmpHome := testutil.InTempHome(t)

	must.MkdirAll("dir1")
	must.MkdirAll("dir2")
	dir1 := filepath.Join(tmpHome, "dir1")
	dir2 := filepath.Join(tmpHome, "dir2")

	Test(t,
		That(`{ tmp pwd = dir1; put $pwd }; put $pwd`).Puts(dir1, tmpHome),
		That(`{ tmp pwd = (num 1); put $pwd }`).Throws(vars.ErrPathMustBeString),
	)

	// We could separate these two test variants into separate unit test
	// modules but that's overkill for this situation and makes the
	// equivalence between the two environments harder to see.
	if runtime.GOOS == "windows" {
		Test(t,
			That(`cd $E:HOME\dir2; { tmp pwd = $E:HOME; put $pwd }; put $pwd`).
				Puts(tmpHome, dir2),
			That(`cd $E:HOME\dir2; { tmp pwd = ..\dir1; put $pwd }; put $pwd`).
				Puts(dir1, dir2),
			That(`cd $E:HOME\dir1; { tmp pwd = ..\dir2; put $pwd }; put $pwd`).
				Puts(dir2, dir1),
		)
	} else {
		Test(t,
			That(`cd ~/dir2; { tmp pwd = ~; put $pwd }; put $pwd`).
				Puts(tmpHome, dir2),
			That(`cd ~/dir2; { tmp pwd = ~/dir1; put $pwd }; put $pwd`).
				Puts(dir1, dir2),
			That(`cd ~/dir1; { tmp pwd = ../dir2; put $pwd }; put $pwd`).
				Puts(dir2, dir1),
		)
	}
}

// Verify the behavior when the CWD cannot be determined.
func TestBuiltinPwd_GetwdError(t *testing.T) {
	testutil.Set(t, Getwd, func() (string, error) { return "", errors.New("cwd unknown") })

	Test(t, That(`put $pwd`).Puts("/unknown/pwd"))
}
