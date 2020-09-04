package eval_test

import (
	"errors"
	"fmt"
	"os/user"
	"path/filepath"
	"testing"

	. "github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/testutil"

	. "github.com/elves/elvish/pkg/eval/evaltest"
	"github.com/elves/elvish/pkg/fsutil"
	"github.com/elves/elvish/pkg/parse"
)

// For error injection into the fsutil.GetHome function.
func currentUser() (*user.User, error) {
	return nil, fmt.Errorf("user unknown")
}

func TestBuiltinFnFS(t *testing.T) {
	tmpHome, cleanup := testutil.InTempHome()
	defer cleanup()

	testutil.MustMkdirAll("dir")
	testutil.MustCreateEmpty("file")

	Test(t,
		That(`path-base a/b/c.png`).Puts("c.png"),
		That("tilde-abbr "+parse.Quote(filepath.Join(tmpHome, "foobar"))).
			Puts(filepath.Join("~", "foobar")),

		That(`-is-dir ~/dir`).Puts(true),
		That(`-is-dir ~/file`).Puts(false),
	)
}

func TestBuiltinCd(t *testing.T) {
	tmpHome, cleanup := testutil.InTempHome()
	defer cleanup()

	testutil.MustMkdirAll("d1")
	d1Path := filepath.Join(tmpHome, "d1")

	// We install this mock for all tests, not just the one that needs it,
	// because it should not be invoked by any of the other tests.
	fsutil.CurrentUser = currentUser
	defer func() { fsutil.CurrentUser = user.Current }()

	Test(t,
		That(`cd dir1 dir2`).Throws(ErrArgs, "cd dir1 dir2"),
		// Basic `cd` test and verification that `$pwd` is correct.
		That(`old = $pwd; cd `+d1Path+`; put $pwd; cd $old; eq $old $pwd`).Puts(d1Path, true),
		// Verify that `cd` with no arg defaults to the home directory.
		That(`cd `+d1Path+`; cd; eq $pwd $E:HOME`).Puts(true),
		// Verify that `cd` with no arg and no $E:HOME var fails since our
		// currentUser mock should result in being unable to dynamically
		// determine the user's home directory.
		That(`unset-env HOME; cd; set-env HOME `+tmpHome).Throws(
			errors.New("can't resolve ~: user unknown"), "cd"),
	)
}

func TestBuiltinDirHistory(t *testing.T) {
	// TODO: Add a Store mock so we can test the behavior when a history Store
	// is available.
	Test(t,
		That(`dir-history`).Throws(ErrStoreNotConnected, "dir-history"),
	)
}
