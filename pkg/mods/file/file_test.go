package file

import (
	"os"
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/testutil"
)

// A number that exceeds the range of int64
const z = "100000000000000000000"

func TestFile(t *testing.T) {
	setup := func(ev *eval.Evaler) {
		ev.ExtendGlobal(eval.BuildNs().AddNs("file", Ns))
	}
	testutil.InTempDir(t)

	TestWithSetup(t, setup,
		That(`
			echo haha > out3
			var f = (file:open out3)
			slurp < $f
			file:close $f
		`).Puts("haha\n"),

		That(`
			var p = (file:pipe)
			echo haha > $p
			file:close $p[w]
			slurp < $p
			file:close $p[r]
		`).Puts("haha\n"),

		That(`
			var p = (file:pipe)
			echo Legolas > $p
			file:close $p[r]
			slurp < $p
		`).Throws(ErrorWithType(&os.PathError{})),

		// Verify that input redirection from a closed pipe throws an exception. That exception is a
		// Go stdlib error whose stringified form looks something like "read |0: file already
		// closed".
		That(`var p = (file:pipe)`, `echo Legolas > $p`, `file:close $p[r]`,
			`slurp < $p`).Throws(ErrorWithType(&os.PathError{})),

		// Side effect checked below
		That("echo > file100", "file:truncate file100 100").DoesNothing(),

		// Should also test the case where the argument doesn't fit in an int
		// but does in a *big.Int, but this could consume too much disk

		That("file:truncate bad -1").Throws(errs.OutOfRange{
			What:     "size argument to file:truncate",
			ValidLow: "0", ValidHigh: "2^64-1", Actual: "-1",
		}),

		That("file:truncate bad "+z).Throws(errs.OutOfRange{
			What:     "size argument to file:truncate",
			ValidLow: "0", ValidHigh: "2^64-1", Actual: z,
		}),

		That("file:truncate bad 1.5").Throws(errs.BadValue{
			What:  "size argument to file:truncate",
			Valid: "integer", Actual: "non-integer",
		}),

		// TODO: Test with PTY when https://b.elv.sh/1595 is resolved.
		That("file:is-tty 0").Puts(false),
		That("file:is-tty (num 0)").Puts(false),
		That(
			"var p = (file:pipe)",
			"file:is-tty $p[r]; file:is-tty $p[w]",
			"file:close $p[r]; file:close $p[w]").
			Puts(false, false),
		That("file:is-tty a").
			Throws(errs.BadValue{What: "argument to file:is-tty",
				Valid: "file value or numerical FD", Actual: "a"}),
		That("file:is-tty []").
			Throws(errs.BadValue{What: "argument to file:is-tty",
				Valid: "file value or numerical FD", Actual: "[]"}),
	)

	fi, err := os.Stat("file100")
	if err != nil {
		t.Errorf("stat file100: %v", err)
	}
	if size := fi.Size(); size != 100 {
		t.Errorf("got file100 size %v, want 100", size)
	}
}
