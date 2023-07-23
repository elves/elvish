package file

import (
	"os"
	"testing"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/testutil"
)

// A number that exceeds the range of int64
const z = "100000000000000000000"

func TestOpen(t *testing.T) {
	testutil.InTempDir(t)
	evaltest.TestWithEvalerSetup(t, setupFileModule,
		That(`
			echo haha > out3
			var f = (file:open out3)
			slurp < $f
			file:close $f
		`).Puts("haha\n"),
	)
}

func TestOpenOutput(t *testing.T) {
	// Run every subtest in a dedicated temporary directory so that they are not
	// affected by files created by previous subtest.
	setup := func(t *testing.T, ev *eval.Evaler) {
		testutil.InTempDir(t)
		setupFileModule(ev)
	}
	evaltest.TestWithSetup(t, setup,
		// &also-input=$true
		That(`
			print foo > file
			var f = (file:open-output &also-input &if-exists=update file)
			read-bytes 1 < $f
			print X > $f
			slurp < $f
			file:close $f
			slurp < file
		`).Puts("f", "o", "fXo"),

		// &if-not-exists=create
		That(`
			var f = (file:open-output new &if-not-exists=create)
			file:close $f
			slurp < new
		`).Puts(""),
		// &if-not-exists=error
		That(`
			var f = (file:open-output new &if-not-exists=error)
		`).Throws(evaltest.ErrorWithType(&os.PathError{})),
		// Default is &if-not-exists=create
		That(`
			var f = (file:open-output new)
			file:close $f
			slurp < new
		`).Puts(""),
		// Invalid &if-not-exists
		That(`
			var f = (file:open-output new &if-not-exists=bad)
		`).Throws(errs.BadValue{What: "if-not-exists option",
			Valid: "create or error", Actual: "bad"}),

		// &if-exists=truncate
		That(`
			print old-content > old
			var f = (file:open-output old &if-exists=truncate)
			print new > $f
			file:close $f
			slurp < old
		`).Puts("new"),
		// &if-exists=append
		That(`
			print old-content > old
			var f = (file:open-output old &if-exists=append)
			print new > $f
			file:close $f
			slurp < old
		`).Puts("old-contentnew"),
		// &if-exists=update
		That(`
			print old-content > old
			var f = (file:open-output old &if-exists=update)
			print new > $f
			file:close $f
			slurp < old
		`).Puts("new-content"),
		// &if-exists=error
		That(`
			print old-content > old
			var f = (file:open-output old &if-exists=error)
		`).Throws(evaltest.ErrorWithType(&os.PathError{})),
		// Default is &if-exists=truncate
		That(`
			print old-content > old
			var f = (file:open-output old)
			print new > $f
			file:close $f
			slurp < old
		`).Puts("new"),
		// Invalid &if-exists
		That(`
			var f = (file:open-output old &if-exists=bad)
		`).Throws(errs.BadValue{What: "if-exists option",
			Valid: "truncate, append, update or error", Actual: "bad"}),
		// &if-exists=error with &if-not-exists=error is an error
		That(`
			var f = (file:open-output old &if-not-exists=error &if-exists=error)
		`).Throws(evaltest.ErrorWithMessage("both &if-not-exists and &if-exists are error")),

		// Invalid &create-perm
		That(`file:open-output new &create-perm=0o1000`).
			Throws(errs.OutOfRange{What: "create-perm option",
				ValidLow: "0", ValidHigh: "0o777", Actual: "0o1000"}),
	)
}

func TestPipe(t *testing.T) {
	evaltest.TestWithEvalerSetup(t, setupFileModule,
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
		`).Throws(evaltest.ErrorWithType(&os.PathError{})),

		// Verify that input redirection from a closed pipe throws an exception. That exception is a
		// Go stdlib error whose stringified form looks something like "read |0: file already
		// closed".
		That(`var p = (file:pipe)`, `echo Legolas > $p`, `file:close $p[r]`,
			`slurp < $p`).Throws(evaltest.ErrorWithType(&os.PathError{})),
	)
}

func TestSeek(t *testing.T) {
	setup := func(t *testing.T, ev *eval.Evaler) {
		testutil.InTempDir(t)
		setupFileModule(ev)
	}
	read1SeekRead1 := func(seekCmd string) string {
		return `
			print 0123456789 > file
			var f = (file:open file)
			read-bytes 1 < $f | nop (all)
			` + seekCmd + `
			read-bytes 1 < $f
			file:close $f
			`
	}
	evaltest.TestWithSetup(t, setup,
		// Default is &whence=start
		That(read1SeekRead1("file:seek $f 1")).Puts("1"),
		// Different &whence
		That(read1SeekRead1("file:seek $f 1 &whence=start")).Puts("1"),
		That(read1SeekRead1("file:seek $f 1 &whence=current")).Puts("2"),
		That(read1SeekRead1("file:seek $f -1 &whence=end")).Puts("9"),
		// TODO: These tests have to close the file before calling file:seek to
		// avoid leaking file descriptors. This shouldn't be necessary.

		// Invalid &whence
		That(read1SeekRead1("file:close $f; file:seek $f -1 &whence=bad")).
			Throws(errs.BadValue{What: "whence",
				Valid: "start, current or end", Actual: "bad"}),
		// Invalid offsets.
		That(read1SeekRead1("file:close $f; file:seek $f "+z)).
			Throws(errs.OutOfRange{What: "offset",
				ValidLow: "-2^64", ValidHigh: "2^64-1", Actual: z}),
		That(read1SeekRead1("file:close $f; file:seek $f 1.5")).
			Throws(errs.BadValue{What: "offset",
				Valid: "exact integer", Actual: "1.5"}),
	)
}

func TestTell(t *testing.T) {
	testutil.InTempDir(t)
	evaltest.TestWithEvalerSetup(t, setupFileModule,
		That(`
			print 0123456789 > file
			var f = (file:open file)
			read-bytes 4 < $f
			file:tell $f
			file:close $f
		`).Puts("0123", 4),
	)
}

func TestTruncate(t *testing.T) {
	testutil.InTempDir(t)
	evaltest.TestWithEvalerSetup(t, setupFileModule,
		// Side effect checked below
		That("echo > file100", "file:truncate file100 100").DoesNothing(),

		// Should also test the case where the argument doesn't fit in an int
		// but does in a *big.Int, but this could consume too much disk

		That("file:truncate bad -1").Throws(errs.OutOfRange{
			What:     "size",
			ValidLow: "0", ValidHigh: "2^64-1", Actual: "-1",
		}),

		That("file:truncate bad "+z).Throws(errs.OutOfRange{
			What:     "size",
			ValidLow: "0", ValidHigh: "2^64-1", Actual: z,
		}),

		That("file:truncate bad 1.5").Throws(errs.BadValue{
			What:  "size",
			Valid: "exact integer", Actual: "1.5",
		}),
	)

	fi, err := os.Stat("file100")
	if err != nil {
		t.Errorf("stat file100: %v", err)
	}
	if size := fi.Size(); size != 100 {
		t.Errorf("got file100 size %v, want 100", size)
	}
}

func TestIsTTY(t *testing.T) {
	evaltest.TestWithEvalerSetup(t, setupFileModule,
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
	if canOpen("/dev/null") {
		evaltest.TestWithEvalerSetup(t, setupFileModule,
			That("file:is-tty 0 < /dev/null").Puts(false),
			That("file:is-tty (num 0) < /dev/null").Puts(false),
		)
	}
	if canOpen("/dev/tty") {
		evaltest.TestWithEvalerSetup(t, setupFileModule,
			That("file:is-tty 0 < /dev/tty").Puts(true),
			That("file:is-tty (num 0) < /dev/tty").Puts(true),
		)
	}
	// TODO: Test with PTY when https://b.elv.sh/1595 is resolved.
}

func canOpen(name string) bool {
	f, err := os.Open(name)
	f.Close()
	return err == nil
}

func setupFileModule(ev *eval.Evaler) {
	ev.ExtendGlobal(eval.BuildNs().AddNs("file", Ns))
}
