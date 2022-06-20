package eval_test

import (
	"errors"
	"path/filepath"
	"testing"

	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/testutil"
)

func TestTildeAbbr(t *testing.T) {
	tmpHome := testutil.InTempHome(t)

	must.MkdirAll("dir")
	must.CreateEmpty("file")

	Test(t,
		That("tilde-abbr "+parse.Quote(filepath.Join(tmpHome, "foobar"))).
			Puts(filepath.Join("~", "foobar")),
	)
}

func TestCd(t *testing.T) {
	tmpHome := testutil.InTempHome(t)

	must.MkdirAll("d1")
	d1Path := filepath.Join(tmpHome, "d1")

	Test(t,
		That(`cd dir1 dir2`).Throws(ErrorWithType(errs.ArityMismatch{}), "cd dir1 dir2"),
		// Basic `cd` test and verification that `$pwd` is correct.
		That("var old = $pwd", "cd "+d1Path, "put $pwd", "cd $old", "eq $old $pwd").
			Puts(d1Path, true),
		// Verify that `cd` with no arg defaults to the home directory.
		That(`cd `+d1Path+`; cd; eq $pwd $E:HOME`).Puts(true),
	)
}

func TestCd_GetHomeError(t *testing.T) {
	err := errors.New("fake error")
	testutil.Set(t, GetHome, func(name string) (string, error) { return "", err })

	Test(t, That("cd").Throws(err, "cd"))
}
