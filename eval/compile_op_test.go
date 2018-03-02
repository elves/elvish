package eval

import "testing"

var opTests = []Test{
	// Chunks
	// ------

	// Empty chunk
	{"", wantNothing},
	// Outputs of pipelines in a chunk are concatenated
	{"put x; put y; put z", want{out: strs("x", "y", "z")}},
	// A failed pipeline cause the whole chunk to fail
	{"put a; e:false; put b", want{out: strs("a"), err: errAny}},

	// Pipelines
	// ---------

	// Pure byte pipeline
	{`echo "Albert\nAllan\nAlbraham\nBerlin" | sed s/l/1/g | grep e`,
		want{bytesOut: []byte("A1bert\nBer1in\n")}},
	// Pure channel pipeline
	{`put 233 42 19 | each [x]{+ $x 10}`, want{out: strs("243", "52", "29")}},
	// Pipeline draining.
	{`range 100 | put x`, want{out: strs("x")}},
	// TODO: Add a useful hybrid pipeline sample

	// Assignments
	// -----------

	// List element assignment
	{"li=[foo bar]; li[0]=233; put $@li",
		want{out: strs("233", "bar")}},
	// Map element assignment
	{"di=[&k=v]; di[k]=lorem; di[k2]=ipsum; put $di[k] $di[k2]",
		want{out: strs("lorem", "ipsum")}},
	{"d=[&a=[&b=v]]; put $d[a][b]; d[a][b]=u; put $d[a][b]",
		want{out: strs("v", "u")}},
	// Multi-assignments.
	{"{a,b}=(put a b); put $a $b", want{out: strs("a", "b")}},
	{"@a=(put a b); put $@a", want{out: strs("a", "b")}},
	{"{a,@b}=(put a b c); put $@b", want{out: strs("b", "c")}},
	//{"di=[&]; di[a b]=(put a b); put $di[a] $di[b]", want{out: strs("a", "b")}},

	// Temporary assignment.
	{"a=alice b=bob; {a,@b}=(put amy ben) put $a $@b; put $a $b",
		want{out: strs("amy", "ben", "alice", "bob")}},
	// Temporary assignment of list element.
	{"l = [a]; l[0]=x put $l[0]; put $l[0]", want{out: strs("x", "a")}},
	// Temporary assignment of map element.
	{"m = [&k=v]; m[k]=v2 put $m[k]; put $m[k]", want{out: strs("v2", "v")}},
	// Temporary assignment before special form.
	{"li=[foo bar] for x $li { put $x }", want{out: strs("foo", "bar")}},

	// Spacey assignment.
	{"a @b = 2 3 foo; put $a $b[1]", want{out: strs("2", "foo")}},
	// Spacey assignment with temporary assignment
	{"x = 1; x=2 y = (+ 1 $x); put $x $y", want{out: strs("1", "3")}},

	// Redirections
	// ------------

	{"f=(mktemp elvXXXXXX); echo 233 > $f; cat < $f; rm $f",
		want{bytesOut: []byte("233\n")}},

	// Redirections from special form.
	{`f = (mktemp elvXXXXXX);
	for x [lorem ipsum] { echo $x } > $f
	cat $f
	rm $f`,
		want{bytesOut: []byte("lorem\nipsum\n")}},

	// Redirections from File object.
	{`fname=(mktemp elvXXXXXX); echo haha > $fname;
			f=(fopen $fname); cat <$f; fclose $f; rm $fname`,
		want{bytesOut: []byte("haha\n")}},

	// Redirections from Pipe object.
	{`p=(pipe); echo haha > $p; pwclose $p; cat < $p; prclose $p`,
		want{bytesOut: []byte("haha\n")}},

	// Redirection from a file closes the value part.
	NewTest(`count < /dev/null`).WantOutStrings("0"),
	NewTest(`fname=(mktemp elvXXXXXX); echo "foo\nbar" > $fname;
			 f=(fopen $fname); count <$f; fclose $f; rm $fname
			`).WantOutStrings("2"),
}

func TestOp(t *testing.T) {
	runTests(t, opTests)
}
