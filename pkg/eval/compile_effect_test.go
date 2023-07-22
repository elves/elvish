package eval_test

import (
	"testing"
	"time"

	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/mods/file"
	"src.elv.sh/pkg/testutil"
)

func TestChunk(t *testing.T) {
	Test(t,
		// Empty chunk
		That("").DoesNothing(),
		// Outputs of pipelines in a chunk are concatenated
		That("put x; put y; put z").Puts("x", "y", "z"),
		// A failed pipeline cause the whole chunk to fail
		That("put a; e:false; put b").Puts("a").Throws(ErrorWithType(ExternalCmdExit{})),
	)
}

func TestPipeline(t *testing.T) {
	Test(t,
		// Pure byte pipeline
		That(`echo "Albert\nAllan\nAlbraham\nBerlin" | sed s/l/1/g | grep e`).
			Prints("A1bert\nBer1in\n"),
		// Pure value pipeline
		That(`put 233 42 19 | each {|x|+ $x 10}`).Puts(243, 52, 29),
		// Pipeline draining.
		That(`range 100 | put x`).Puts("x"),
		// TODO: Add a useful hybrid pipeline sample
	)
}

func TestPipeline_BgJob(t *testing.T) {
	setup := func(ev *Evaler) {
		ev.ExtendGlobal(BuildNs().AddNs("file", file.Ns))
	}

	notes1 := make(chan string)
	notes2 := make(chan string)

	putNote := func(ch chan<- string) func(*Evaler) {
		return func(ev *Evaler) {
			ev.BgJobNotify = func(note string) { ch <- note }
		}
	}
	verifyNote := func(notes <-chan string, wantNote string) func(t *testing.T) {
		return func(t *testing.T) {
			select {
			case note := <-notes:
				if note != wantNote {
					t.Errorf("got note %q, want %q", note, wantNote)
				}
			case <-time.After(testutil.Scaled(100 * time.Millisecond)):
				t.Errorf("timeout waiting for notification")
			}
		}
	}

	TestWithEvalerSetup(t, setup,
		That(
			"set notify-bg-job-success = $false",
			"var p = (file:pipe)",
			"{ print foo > $p; file:close $p[w] }&",
			"slurp < $p; file:close $p[r]").
			Puts("foo"),
		// Notification
		That(
			"set notify-bg-job-success = $true",
			"var p = (file:pipe)",
			"fn f { file:close $p[w] }",
			"f &",
			"slurp < $p; file:close $p[r]").
			Puts("").
			WithSetup(putNote(notes1)).
			Passes(verifyNote(notes1, "job f & finished")),
		// Notification, with exception
		That(
			"set notify-bg-job-success = $true",
			"var p = (file:pipe)",
			"fn f { file:close $p[w]; fail foo }",
			"f &",
			"slurp < $p; file:close $p[r]").
			Puts("").
			WithSetup(putNote(notes2)).
			Passes(verifyNote(notes2, "job f & finished, errors = foo")),
	)
}

func TestPipeline_ReaderGone(t *testing.T) {
	// See Unix-only tests in compile_effect_unix_test.go.
	Test(t,
		// Internal commands writing to byte output raises ReaderGone when the
		// reader is exited, which is then suppressed.
		That("while $true { echo y } | nop").DoesNothing(),
		That(
			"var reached = $false",
			"{ while $true { echo y }; reached = $true } | nop",
			"put $reached",
		).Puts(false),
		// Similar for value output.
		That("while $true { put y } | nop").DoesNothing(),
		That(
			"var reached = $false",
			"{ while $true { put y }; reached = $true } | nop",
			"put $reached",
		).Puts(false),
	)
}

func TestCommand(t *testing.T) {
	Test(t,
		That("put foo").Puts("foo"),
		// Command errors when the head is not a single value.
		That("{put put} foo").Throws(
			errs.ArityMismatch{What: "command",
				ValidLow: 1, ValidHigh: 1, Actual: 2},
			"{put put}"),
		// Command errors when the head is not callable or string containing slash.
		That("[] foo").Throws(
			errs.BadValue{
				What:   "command",
				Valid:  "callable or string containing slash",
				Actual: "[]"},
			"[]"),
		// Command errors when argument errors.
		That("put [][1]").Throws(ErrorWithType(errs.OutOfRange{}), "[][1]"),
		// Command errors when an option key is not string.
		That("put &[]=[]").Throws(
			errs.BadValue{What: "option key", Valid: "string", Actual: "list"},
			"put &[]=[]"),
		// Command errors when any optional evaluation errors.
		That("put &x=[][1]").Throws(ErrorWithType(errs.OutOfRange{}), "[][1]"),
	)
}

func TestCommand_Special(t *testing.T) {
	Test(t,
		// Regression test for #1204; ensures that the arguments of special
		// forms are not accidentally compiled twice.
		That("nop (and (use builtin)); nop $builtin:echo~").DoesNothing(),

		// Behavior of individual special commands are tested in
		// builtin_special_test.go.
	)
}

func TestCommand_LegacyTemporaryAssignment(t *testing.T) {
	Test(t,
		That("var a b = alice bob; {a,@b}=(put amy ben) put $a $@b; put $a $b").
			Puts("amy", "ben", "alice", "bob").PrintsStderrWith("deprecated"),
		// Temporary assignment of list element.
		That("var l = [a]; l[0]=x put $l[0]; put $l[0]").
			Puts("x", "a").PrintsStderrWith("deprecated"),
		// Temporary assignment of map element.
		That("var m = [&k=v]; m[k]=v2 put $m[k]; put $m[k]").
			Puts("v2", "v").PrintsStderrWith("deprecated"),
		// Temporary assignment before special form.
		That("li=[foo bar] for x $li { put $x }").
			Puts("foo", "bar").PrintsStderrWith("deprecated"),
		// Multiple LHSs in temporary assignments.
		That("{a b}={foo bar} put $a $b").
			Puts("foo", "bar").PrintsStderrWith("deprecated"),
		That("@a=(put a b) put $@a").
			Puts("a", "b").PrintsStderrWith("deprecated"),
		That("{a,@b}=(put a b c) put $@b").
			Puts("b", "c").PrintsStderrWith("deprecated"),
		// Using syntax of temporary assignment for non-temporary assignment no
		// longer compiles
		That("x=y").DoesNotCompile(
			`using the syntax of temporary assignment for non-temporary assignment is no longer supported; use "var" or "set" instead`),
	)
}

func TestCommand_LegacyTemporaryAssignmentSyntaxIsDeprecated(t *testing.T) {
	testCompileTimeDeprecation(t, "a=foo echo $a",
		"the legacy temporary assignment syntax is deprecated", 18)
}

func TestCommand_Redir(t *testing.T) {
	setup := func(ev *Evaler) {
		ev.ExtendGlobal(BuildNs().AddNs("file", file.Ns))
	}
	testutil.InTempDir(t)

	TestWithEvalerSetup(t, setup,
		// Output and input redirection.
		That("echo 233 > out1", " slurp < out1").Puts("233\n"),
		// Append.
		That("echo 1 > out; echo 2 >> out; slurp < out").Puts("1\n2\n"),
		// Read and write.
		// TODO: Add a meaningful use case that uses both read and write.
		That("echo 233 <> out1", " slurp < out1").Puts("233\n"),

		// Redirections from special form.
		That(`for x [lorem ipsum] { echo $x } > out2`, `slurp < out2`).
			Puts("lorem\nipsum\n"),

		// Using numeric FDs as source and destination.
		That(`{ echo foobar >&2 } 2> out3`, `slurp < out3`).
			Puts("foobar\n"),
		// Using named FDs as source and destination.
		That("echo 233 stdout> out1", " slurp stdin< out1").Puts("233\n"),
		That(`{ echo foobar >&stderr } stderr> out4`, `slurp < out4`).
			Puts("foobar\n"),
		// Using a new FD as source throws an exception.
		That(`echo foo >&4`).Throws(InvalidFD{FD: 4}),
		// Using a new FD as destination is OK, and makes it available.
		That(`{ echo foo >&4 } 4>out5`, `slurp < out5`).Puts("foo\n"),

		// Redirections from File object.
		That(`echo haha > out3`, `var f = (file:open out3)`, `slurp <$f`, ` file:close $f`).
			Puts("haha\n"),
		// Redirections from Pipe object.
		That(`var p = (file:pipe); echo haha > $p; file:close $p[w]; slurp < $p; file:close $p[r]`).
			Puts("haha\n"),

		// We can't read values from a file and shouldn't hang when iterating
		// over input from a file.
		// Regression test for https://src.elv.sh/issues/1010
		That("echo abc > bytes", "each $echo~ < bytes").Prints("abc\n"),
		That("echo def > bytes", "only-values < bytes | count").Puts(0),

		// Writing value output to file throws an exception.
		That("put foo >a").Throws(ErrPortDoesNotSupportValueOutput, "put foo >a"),
		// Writing value output to closed port throws an exception too.
		That("put foo >&-").Throws(ErrPortDoesNotSupportValueOutput, "put foo >&-"),

		// Invalid redirection destination.
		That("echo []> test").Throws(
			errs.BadValue{
				What:  "redirection destination",
				Valid: "fd name or number", Actual: "[]"},
			"[]"),
		// Invalid fd redirection source.
		That("echo >&test").Throws(
			errs.BadValue{
				What:  "redirection source",
				Valid: "fd name or number or '-'", Actual: "test"},
			"test"),
		// Invalid redirection source.
		That("echo > []").Throws(
			errs.BadValue{
				What:  "redirection source",
				Valid: "string, file or map", Actual: "list"},
			"[]"),
		// Invalid map for redirection.
		That("echo < [&]").Throws(
			errs.BadValue{
				What:  "map for input redirection",
				Valid: "map with file in the 'r' field", Actual: "[&]"},
			"[&]"),
		That("echo > [&]").Throws(
			errs.BadValue{
				What:  "map for output redirection",
				Valid: "map with file in the 'w' field", Actual: "[&]"},
			"[&]"),

		// Exception when evaluating source or destination.
		That("echo > (fail foo)").Throws(FailError{"foo"}, "fail foo"),
		That("echo (fail foo)> file").Throws(FailError{"foo"}, "fail foo"),
	)
}

func TestCommand_Stacktrace(t *testing.T) {
	oops := ErrorWithMessage("oops")
	Test(t,
		// Stack traces.
		That("fail oops").Throws(oops, "fail oops"),
		That("fn f { fail oops }", "f").Throws(oops, "fail oops ", "f"),
		That("fn f { fail oops }", "fn g { f }", "g").Throws(
			oops, "fail oops ", "f ", "g"),
		// Error thrown before execution.
		That("fn f { }", "f a").Throws(ErrorWithType(errs.ArityMismatch{}), "f a"),
		// Error from builtin.
		That("count 1 2 3").Throws(
			ErrorWithType(errs.ArityMismatch{}), "count 1 2 3"),
	)
}
