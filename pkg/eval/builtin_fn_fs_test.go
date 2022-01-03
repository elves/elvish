package eval_test

import (
	"errors"
	"fmt"
	"os/user"
	"path/filepath"
	"testing"

	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/testutil"
)

// For error injection into the fsutil.GetHome function.
func currentUser() (*user.User, error) {
	return nil, fmt.Errorf("user unknown")
}

func TestTildeAbbr(t *testing.T) {
	tmpHome := testutil.InTempHome(t)

	testutil.MustMkdirAll("dir")
	testutil.MustCreateEmpty("file")

	Test(t,
		That("tilde-abbr "+parse.Quote(filepath.Join(tmpHome, "foobar"))).
			Puts(filepath.Join("~", "foobar")),
	)
}

func TestCd(t *testing.T) {
	tmpHome := testutil.InTempHome(t)

	testutil.MustMkdirAll("d1")
	d1Path := filepath.Join(tmpHome, "d1")

	// We install this mock for all tests, not just the one that needs it,
	// because it should not be invoked by any of the other tests.
	fsutil.CurrentUser = currentUser
	defer func() { fsutil.CurrentUser = user.Current }()

	Test(t,
		That(`cd dir1 dir2`).Throws(ErrorWithType(errs.ArityMismatch{}), "cd dir1 dir2"),
		// Basic `cd` test and verification that `$pwd` is correct.
		That("var old = $pwd", "cd "+d1Path, "put $pwd", "cd $old", "eq $old $pwd").
			Puts(d1Path, true),
		// Verify that `cd` with no arg defaults to the home directory.
		That(`cd `+d1Path+`; cd; eq $pwd $E:HOME`).Puts(true),
		// Verify that `cd` with no arg and no $E:HOME var fails since our
		// currentUser mock should result in being unable to dynamically
		// determine the user's home directory.
		That(`unset-env HOME; cd; set-env HOME `+tmpHome).Throws(
			errors.New("can't resolve ~: user unknown"), "cd"),
	)
}
