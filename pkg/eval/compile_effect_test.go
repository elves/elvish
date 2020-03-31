package eval

import (
	"testing"

	"github.com/elves/elvish/pkg/util"
)

func TestCompileEffect(t *testing.T) {
	_, cleanup := util.InTestDir()
	defer cleanup()

	Test(t,
		// Chunks
		// ------

		// Empty chunk
		That("").DoesNothing(),
		// Outputs of pipelines in a chunk are concatenated
		That("put x; put y; put z").Puts("x", "y", "z"),
		// A failed pipeline cause the whole chunk to fail
		That("put a; e:false; put b").Puts("a").ThrowsAny(),

		// Pipelines
		// ---------

		// Pure byte pipeline
		That(`echo "Albert\nAllan\nAlbraham\nBerlin" | sed s/l/1/g | grep e`).
			Prints("A1bert\nBer1in\n"),
		// Pure channel pipeline
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

		// Commands
		// --------

		That("put foo").Puts("foo"),
		// Command errors when the head is not a single value.
		That("{put put} foo").ThrowsMessage("head of command must be a single value; got 2 values"),
		// Command errors when when argument errors.
		That("put [][1]").ThrowsMessage("index out of range"),
		// Command errors when any optional evaluation errors.
		That("put &x=[][1]").ThrowsMessage("index out of range"),

		// Assignments
		// -----------

		// Spacey assignment.
		That("a @b = 2 3 foo; put $a $b[1]").Puts("2", "foo"),
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

		// Assignment errors when the RHS errors.
		That("x = [][1]").ThrowsMessage("index out of range"),
		// Arity mismatch.
		That("x = 1 2").Throws(ErrArityMismatch),
		That("x y = 1").Throws(ErrArityMismatch),
		That("x y @z = 1").Throws(ErrArityMismatch),

		// Redirections
		// ------------

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
		That(`echo foo >&4`).ThrowsAny(),
		// Using a new FD as destination is OK, and makes it available.
		That(`{ echo foo >&4 } 4>out5`, `slurp < out5`).Puts("foo\n"),

		// Redirections from File object.
		That(`echo haha > out3`, `f = (fopen out3)`, `slurp <$f`, ` fclose $f`).
			Puts("haha\n"),
		// Redirections from Pipe object.
		That(`p = (pipe); echo haha > $p; pwclose $p; slurp < $p; prclose $p`).
			Puts("haha\n"),

		// Using anything else in redirection throws an exception.
		That("echo > []").ThrowsMessage("redirection source must be string, file or pipe; got list"),
	)
}
