//go:build unix

package unix

import (
	"os"
	"strconv"
	"sync"
	"testing"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/testutil"
)

func TestUmask(t *testing.T) {
	testutil.Umask(t, 0o22)

	evaltest.TestWithEvalerSetup(t, useUnix,
		// Start with a known umask value. Note that we can't rely on the umask
		// set at the start of the test because the internally cached umask
		// value is only set during init and won't pick up the value in the
		// test.
		That(`set unix:umask = 022`).DoesNothing(),

		// Read the variable.
		That(`put $unix:umask`).Puts(`0o022`),
		// A string with no base prefix is parsed as octal.
		That(`set unix:umask = 23; put $unix:umask`).Puts(`0o023`),
		// A string with 0o is accepted too.
		That(`set unix:umask = 0o75; put $unix:umask`).Puts(`0o075`),
		// People won't normally use non-octal bases but make sure these cases
		// behave sensibly given that Elvish supports number literals with an
		// explicit base.
		That(`set unix:umask = 0x43; put $unix:umask`).Puts("0o103"),
		That(`set unix:umask = 0b001010100; put $unix:umask`).Puts("0o124"),
		// An exact integer is supported.
		That(`set unix:umask = (num 0o123); put $unix:umask`).Puts(`0o123`),
		// An inexact integer is supported too.
		That(`set unix:umask = (num 9.0); put $unix:umask`).Puts(`0o011`),

		// Test that setting the umask affects external commands.
		//
		// The output of umask is unspecified in POSIX, but all Unix flavors
		// Elvish supports write a 0 followed by an octal number. There is one
		// inconsistency though: OpenBSD does not zero-pad the number (other
		// than the leading 0), so a umask of 0o012 will appear as 012 on
		// OpenBSD but 0012 on other platforms. Avoid this by using a umask that
		// is 3 octal digits long.
		That(`set unix:umask = 0123; sh -c umask`).Prints("0123\n"),

		// Temporarily assigning $unix:umask.
		That(
			`set unix:umask = 022`,
			`{ tmp unix:umask = 011; put $unix:umask }`,
			`put $unix:umask`).Puts("0o011", "0o022"),

		// Parse errors when setting unix:umask.

		// A fractional inexact number.
		That(`set unix:umask = (num 123.4)`).Throws(
			errs.BadValue{What: "umask", Valid: validUmaskMsg, Actual: "123.4"}),

		// A string that can't be parsed as a number at all.
		That(`set unix:umask = 022z`).Throws(errs.BadValue{
			What: "umask", Valid: validUmaskMsg, Actual: "022z"}),

		// Invalid type.
		That(`set unix:umask = [1]`).Throws(errs.BadValue{
			What: "umask", Valid: validUmaskMsg, Actual: "list"}),

		// Values outside the legal range.
		That(`set unix:umask = 0o1000`).Throws(errs.OutOfRange{
			What: "umask", ValidLow: "0", ValidHigh: "0o777", Actual: "0o1000"}),
		That(`set unix:umask = -1`).Throws(errs.OutOfRange{
			What: "umask", ValidLow: "0", ValidHigh: "0o777", Actual: "-0o1"}),
	)
}

func TestUmaskGetRace(t *testing.T) {
	testutil.Umask(t, 0o22)
	testutil.InTempDir(t)

	// An old implementation of UmaskVariable.Get had a bug where it will
	// briefly set the umask to 0, which can impact files created concurrently.
	for i := 0; i < 100; i++ {
		filename := strconv.Itoa(i)
		runParallel(
			func() {
				// Calling UmaskVariable.Get is much quicker than creating a
				// file, so do it in a loop to increase the chance of triggering
				// a race condition.
				for j := 0; j < 100; j++ {
					UmaskVariable{}.Get()
				}
			},
			func() {
				must.OK(create(filename, 0o666))
			})

		perm := must.OK1(os.Stat(filename)).Mode().Perm()
		if perm != 0o644 {
			t.Errorf("got perm %o, want 0o644 (run %d)", perm, i)
		}
	}
}

func runParallel(funcs ...func()) {
	var wg sync.WaitGroup
	wg.Add(len(funcs))
	for _, f := range funcs {
		f := f
		go func() {
			f()
			wg.Done()
		}()
	}
	wg.Wait()
}

func create(name string, perm os.FileMode) error {
	file, err := os.OpenFile(name, os.O_CREATE, perm)
	if err == nil {
		file.Close()
	}
	return err
}
