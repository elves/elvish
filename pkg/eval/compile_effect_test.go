package eval

import "testing"

func TestCompileEffect(t *testing.T) {
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

		// Assignments
		// -----------

		// List element assignment
		That("li=[foo bar]; li[0]=233; put $@li").Puts("233", "bar"),
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

		// Spacey assignment.
		That("a @b = 2 3 foo; put $a $b[1]").Puts("2", "foo"),
		// Spacey assignment with temporary assignment
		That("x = 1; x=2 y = (+ 1 $x); put $x $y").Puts("1", 3.0),

		// Redirections
		// ------------

		That("f=(mktemp elvXXXXXX); echo 233 > $f; cat < $f; rm $f").
			Prints("233\n"),

		// Redirections from special form.
		That(`f = (mktemp elvXXXXXX);
	for x [lorem ipsum] { echo $x } > $f
	cat $f
	rm $f`).Prints("lorem\nipsum\n"),

		// Redirections from File object.
		That(`fname=(mktemp elvXXXXXX); echo haha > $fname;
			f=(fopen $fname); cat <$f; fclose $f; rm $fname`).Prints("haha\n"),

		// Redirections from Pipe object.
		That(`p=(pipe); echo haha > $p; pwclose $p; cat < $p; prclose $p`).
			Prints("haha\n"),
	)
}
