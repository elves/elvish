package eval_test

import (
	"testing"

	"src.elv.sh/pkg/eval/errs"

	. "src.elv.sh/pkg/eval"
	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/testutil"
)

func TestChunk(t *testing.T) {
	Test(t,
		// Empty chunk
		That("").DoesNothing(),
		// Outputs of pipelines in a chunk are concatenated
		That("put x; put y; put z").Puts("x", "y", "z"),
		// A failed pipeline cause the whole chunk to fail
		That("put a; e:false; put b").Puts("a").Throws(AnyError),
	)
}

func TestPipeline(t *testing.T) {
	Test(t,
		// Pure byte pipeline
		That(`echo "Albert\nAllan\nAlbraham\nBerlin" | sed s/l/1/g | grep e`).
			Prints("A1bert\nBer1in\n"),
		// Pure value pipeline
		That(`put 233 42 19 | each [x]{+ $x 10}`).Puts(243.0, 52.0, 29.0),
		// Pipeline draining.
		That(`range 100 | put x`).Puts("x"),
		// Background pipeline.
		That(
			"notify-bg-job-success = $false",
			"p = (pipe)",
			"{ print foo > $p; pwclose $p }&",
			"slurp < $p",
			"prclose $p").Puts("foo"),
		// TODO: Add a useful hybrid pipeline sample
	)
}

func TestCommand(t *testing.T) {
	Test(t,
		That("put foo").Puts("foo"),
		// Command errors when the head is not a single value.
		That("{put put} foo").Throws(
			errs.ArityMismatch{
				What: "command", ValidLow: 1, ValidHigh: 1, Actual: 2},
			"{put put}"),
		// Command errors when the head is not callable or string containing slash.
		That("[] foo").Throws(
			errs.BadValue{
				What:   "command",
				Valid:  "callable or string containing slash",
				Actual: "list"},
			"[]"),
		// Command errors when when argument errors.
		That("put [][1]").Throws(ErrorWithType(errs.OutOfRange{}), "[][1]"),
		// Command errors when an option key is not string.
		That("put &[]=[]").Throws(
			errs.BadValue{What: "option key", Valid: "string", Actual: "list"},
			"put &[]=[]"),
		// Command errors when any optional evaluation errors.
		That("put &x=[][1]").Throws(ErrorWithType(errs.OutOfRange{}), "[][1]"),
	)
}

func TestCommand_DeprecateUsingTemporaryAssignmentForNonTemporaryAssignment(t *testing.T) {
	testCompileTimeDeprecation(t, "foo=bar", `using the syntax of temporary assignment for non-temporary assignment is deprecated; use "var" or "set" instead`, 15)

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

func TestCommand_Assignment(t *testing.T) {
	// NOTE: TestClosure has more tests for the interaction between assignment
	// and variable scoping.

	Test(t,
		// Spacey assignment.
		That("a = foo; put $a").Puts("foo"),
		That("a b = foo bar; put $a $b").Puts("foo", "bar"),
		That("a @b = 2 3 foo; put $a $b").Puts("2", vals.MakeList("3", "foo")),
		That("a @b c = 1 2 3 4; put $a $b $c").
			Puts("1", vals.MakeList("2", "3"), "4"),
		That("a @b c = 1 2; put $a $b $c").Puts("1", vals.EmptyList, "2"),
		That("@a = ; put $a").Puts(vals.EmptyList),

		// List element assignment
		That("li=[foo bar]; li[0]=233; put $@li").Puts("233", "bar"),
		// Variable in list assignment must already be defined. Regression test
		// for b.elv.sh/889.
		That("foobarlorem[0] = a").DoesNotCompile(),
		// Map element assignment
		That("di=[&k=v]; di[k]=lorem; di[k2]=ipsum; put $di[k] $di[k2]").
			Puts("lorem", "ipsum"),
		That("d=[&a=[&b=v]]; put $d[a][b]; d[a][b]=u; put $d[a][b]").
			Puts("v", "u"),
		// Multi-assignments.
		That("{a,b}=(put a b); put $a $b").Puts("a", "b"),
		That("@a=(put a b); put $@a").Puts("a", "b"),
		That("{a,@b}=(put a b c); put $@b").Puts("b", "c"),
		//That("di=[&]; di[a b]=(put a b); put $di[a] $di[b]").Puts("a", "b"),

		// Temporary assignment.
		That("a=alice b=bob; {a,@b}=(put amy ben) put $a $@b; put $a $b").
			Puts("amy", "ben", "alice", "bob"),
		// Temporary assignment of list element.
		That("l = [a]; l[0]=x put $l[0]; put $l[0]").Puts("x", "a"),
		// Temporary assignment of map element.
		That("m = [&k=v]; m[k]=v2 put $m[k]; put $m[k]").Puts("v2", "v"),
		// Temporary assignment before special form.
		That("li=[foo bar] for x $li { put $x }").Puts("foo", "bar"),
		// Spacey assignment with temporary assignment
		That("x = 1; x=2 y = (+ 1 $x); put $x $y").Puts("1", 3.0),

		// Concurrently creating a new variable and accessing existing variable.
		// Run with "go test -race".
		That("x = 1", "put $x | y = (all)").DoesNothing(),
		That("nop (x = 1) | nop").DoesNothing(),

		// Assignment errors when the RHS errors.
		That("x = [][1]").Throws(ErrorWithType(errs.OutOfRange{}), "[][1]"),
		// Assignment errors itself.
		That("true = 1").Throws(vars.ErrSetReadOnlyVar, "true"),
		That("@true = 1").Throws(vars.ErrSetReadOnlyVar, "@true"),
		// A readonly var as a target for the `except` clause should error.
		That("try { fail reason } except nil { }").Throws(vars.ErrSetReadOnlyVar, "nil"),
		That("try { fail reason } except x { }").DoesNothing(),
		// Evaluation of the assignability occurs at run-time so, if no exception is raised, this
		// otherwise invalid use of `nil` is okay.
		That("try { } except nil { }").DoesNothing(),
		// Arity mismatch.
		That("x = 1 2").Throws(
			errs.ArityMismatch{
				What:     "assignment right-hand-side",
				ValidLow: 1, ValidHigh: 1, Actual: 2},
			"x = 1 2"),
		That("x y = 1").Throws(
			errs.ArityMismatch{
				What:     "assignment right-hand-side",
				ValidLow: 2, ValidHigh: 2, Actual: 1},
			"x y = 1"),
		That("x y @z = 1").Throws(
			errs.ArityMismatch{
				What:     "assignment right-hand-side",
				ValidLow: 2, ValidHigh: -1, Actual: 1},
			"x y @z = 1"),

		// Trying to add a new name in a namespace throws an exception.
		// Regression test for #1214.
		That("ns: = (ns [&]); ns:a = b").Throws(NoSuchVariable("ns:a"), "ns:a = b"),
	)
}

func TestCommand_Redir(t *testing.T) {
	_, cleanup := testutil.InTestDir()
	defer cleanup()

	Test(t,
		// Output and input redirection.
		That("echo 233 > out1", " slurp < out1").Puts("233\n"),
		// Append.
		That("echo 1 > out; echo 2 >> out; slurp < out").Puts("1\n2\n"),

		// Redirections from special form.
		That(`for x [lorem ipsum] { echo $x } > out2`, `slurp < out2`).
			Puts("lorem\nipsum\n"),

		// Using numeric FDs as source and destination.
		That(`{ echo foobar >&2 } 2> out3`, `slurp < out3`).
			Puts("foobar\n"),
		// Using named FDs as source and destination.
		That(`{ echo foobar >&stderr } stderr> out4`, `slurp < out4`).
			Puts("foobar\n"),
		// Using a new FD as source throws an exception.
		That(`echo foo >&4`).Throws(AnyError),
		// Using a new FD as destination is OK, and makes it available.
		That(`{ echo foo >&4 } 4>out5`, `slurp < out5`).Puts("foo\n"),

		// Redirections from File object.
		That(`echo haha > out3`, `f = (fopen out3)`, `slurp <$f`, ` fclose $f`).
			Puts("haha\n"),
		// Redirections from Pipe object.
		That(`p = (pipe); echo haha > $p; pwclose $p; slurp < $p; prclose $p`).
			Puts("haha\n"),

		// We can't read values from a file and shouldn't hang when iterating
		// over input from a file.
		// Regression test for https://src.elv.sh/issues/1010
		That("echo abc > bytes", "each $echo~ < bytes").Prints("abc\n"),
		That("echo def > bytes", "only-values < bytes | count").Puts("0"),

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
				Valid: "string, file or pipe", Actual: "list"},
			"[]"),
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
