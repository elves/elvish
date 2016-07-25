package eval

import (
	"errors"
	"os"
	"reflect"
	"strconv"
	"syscall"
	"testing"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

func TestBuiltinNamespace(t *testing.T) {
	pid := strconv.Itoa(syscall.Getpid())
	if ToString(builtinNamespace["pid"].Get()) != pid {
		t.Errorf(`ev.builtin["pid"] = %v, want %v`, builtinNamespace["pid"], pid)
	}
}

var errAny = errors.New("")

type more struct {
	wantBytesOut []byte
	wantError    error
}

var nomore more

var evalTests = []struct {
	text    string
	wantOut []Value
	more
}{
	// Chunks.
	// Empty chunk
	{"", []Value{}, nomore},
	// Outputs of pipelines in a chunk are concatenated
	{"put x; put y; put z", strs("x", "y", "z"), nomore},
	// A failed pipeline cause the whole chunk to fail
	{"put a; e:false; put b", strs("a"), more{
		wantError: &util.PosError{7, 14, FakeExternalCmdExit("false", 1, 0)}}},

	// Pipelines.
	// Pure byte pipeline
	{`echo "Albert\nAllan\nAlbraham\nBerlin" | sed s/l/1/g | grep e`,
		[]Value{}, more{wantBytesOut: []byte("A1bert\nBer1in\n")}},
	// Pure channel pipeline
	{`put 233 42 19 | each [x]{+ $x 10}`, strs("243", "52", "29"), nomore},
	// TODO: Add a useful hybrid pipeline sample

	// List element assignment
	{"li=[foo bar]; li[0]=233; put $@li", strs("233", "bar"), nomore},
	// Map element assignment
	{"di=[&k=v]; di[k]=lorem; di[k2]=ipsum; put $di[k] $di[k2]",
		strs("lorem", "ipsum"), nomore},
	{"d=[&a=[&b=v]]; put $d[a][b]; d[a][b]=u; put $d[a][b]",
		strs("v", "u"), nomore},
	// Multi-assignments.
	{"{a,b}=`put a b`; put $a $b", strs("a", "b"), nomore},
	{"@a=`put a b`; put $@a", strs("a", "b"), nomore},
	{"{a,@b}=`put a b c`; put $@b", strs("b", "c"), nomore},
	// {"di=[&]; di[a b]=`put a b`; put $di[a] $di[b]", strs("a", "b"), nomore},
	// Temporary assignment.
	{"a=alice b=bob; {a,@b}=(put amy ben) put $a $@b; put $a $b",
		strs("amy", "ben", "alice", "bob"), nomore},

	// Control structures.
	// if
	{"if true; then put then; fi", strs("then"), nomore},
	{"if false; then put then; else put else; fi", strs("else"), nomore},
	{"if false; then put 1; elif false; then put 2; else put 3; fi",
		strs("3"), nomore},
	{"if false; then put 2; elif true; then put 2; else put 3; fi",
		strs("2"), nomore},
	// while
	{"x=0; while < $x 4; do put $x; x=(+ $x 1); done",
		strs("0", "1", "2", "3"), nomore},
	// for
	{"for x in tempora mores; do put 'O '$x; done",
		strs("O tempora", "O mores"), nomore},
	// break
	{"for x in a; do break; else put $x; done", strs(), nomore},
	// else
	{"for x in a; do put $x; else put $x; done", strs("a"), nomore},
	// continue
	{"for x in a b; do put $x; continue; put $x; done", strs("a", "b"), nomore},
	// begin/end
	{"begin; put lorem; put ipsum; end", strs("lorem", "ipsum"), nomore},

	// Redirections.
	{"f=`mktemp elvXXXXXX`; echo 233 > $f; cat < $f; rm $f", strs(),
		more{wantBytesOut: []byte("233\n")}},
	// Redirections from File object.
	{`fname=(mktemp elvXXXXXX); echo haha > $fname;
	f=(fopen $fname); cat <$f; fclose $f; rm $fname`, strs(),
		more{wantBytesOut: []byte("haha\n")}},
	// Redirections from Pipe object.
	{`p=(pipe); echo haha > $p; pwclose $p; cat < $p; prclose $p`, strs(),
		more{wantBytesOut: []byte("haha\n")}},

	// Compounding.
	{"put {fi,elvi}sh{1.0,1.1}",
		strs("fish1.0", "fish1.1", "elvish1.0", "elvish1.1"), nomore},

	// List, map and indexing
	{"println [a b c] [&key=value] | each put",
		strs("[a b c] [&key=value]"), nomore},
	{"put [a b c][2]", strs("c"), nomore},
	{"put [;a;b c][2][0]", strs("b"), nomore},
	{"put [&key=value][key]", strs("value"), nomore},

	// String literal
	{`put 'such \"''literal'`, strs(`such \"'literal`), nomore},
	{`put "much \n\033[31;1m$cool\033[m"`,
		strs("much \n\033[31;1m$cool\033[m"), nomore},

	// Output capture
	{"put (put lorem ipsum)", strs("lorem", "ipsum"), nomore},

	// Boolean capture
	{"put ?(true) ?(false)",
		[]Value{Bool(true), Bool(false)}, nomore},

	// Variable and compounding
	{"x='SHELL'\nput 'WOW, SUCH '$x', MUCH COOL'\n",
		strs("WOW, SUCH SHELL, MUCH COOL"), nomore},
	// Splicing
	{"x=[elvish rules]; put $@x", strs("elvish", "rules"), nomore},

	// Wildcard.
	{"put /*", strs(util.RootStar()...), nomore},

	// Tilde.
	{"h=$env:HOME; env:HOME=/foo; put ~ ~/src; env:HOME=$h",
		strs("/foo", "/foo/src"), nomore},

	// Closure
	// Basics
	{"[]{ }", strs(), nomore},
	{"[x]{put $x} foo", strs("foo"), nomore},
	// Variable capture
	{"x=lorem; []{x=ipsum}; put $x", strs("ipsum"), nomore},
	{"x=lorem; []{ put $x; x=ipsum }; put $x",
		strs("lorem", "ipsum"), nomore},
	// Shadowing
	{"x=ipsum; []{ local:x=lorem; put $x }; put $x",
		strs("lorem", "ipsum"), nomore},
	// Shadowing by argument
	{"x=ipsum; [x]{ put $x; x=BAD } lorem; put $x",
		strs("lorem", "ipsum"), nomore},
	// Closure captures new local variables every time
	{`fn f []{ x=0; put []{x=(+ $x 1)} []{put $x} }
      {inc1,put1}=(f); $put1; $inc1; $put1
	  {inc2,put2}=(f); $put2; $inc2; $put2`,
		strs("0", "1", "0", "1"), nomore},
	// Positional variables.
	{`{ put $1 } lorem ipsum`, strs("ipsum"), nomore},
	// Positional variables in the up: namespace.
	{`{ { put $up:0 } in } out`, strs("out"), nomore},

	// fn.
	{"fn f [x]{ put x=$x'.' }; f lorem; f ipsum",
		strs("x=lorem.", "x=ipsum."), nomore},
	// return.
	{"fn f []{ put a; return; put b }; f", strs("a"), nomore},

	// rest args and $args.
	{"[x @xs]{ put $x $xs $args } a b c",
		[]Value{String("a"),
			NewList(String("b"), String("c")),
			NewList(String("a"), String("b"), String("c"))}, nomore},
	// $args.
	{"{ put $args } lorem ipsum",
		[]Value{NewList(String("lorem"), String("ipsum"))}, nomore},

	// Namespaces
	// Pseudo-namespaces local: and up:
	{"x=lorem; []{local:x=ipsum; put $up:x $local:x}",
		strs("lorem", "ipsum"), nomore},
	{"x=lorem; []{up:x=ipsum; put $x}; put $x",
		strs("ipsum", "ipsum"), nomore},
	// Pseudo-namespace env:
	{"env:foo=lorem; put $env:foo", strs("lorem"), nomore},
	{"del env:foo; put $env:foo", strs(""), nomore},
	// TODO: Test module namespace

	// Builtin functions
	// Arithmetics
	// TODO test more edge cases
	{"+ 233100 233", strs("233333"), nomore},
	{"- 233333 233100", strs("233"), nomore},
	{"* 353 661", strs("233333"), nomore},
	{"/ 233333 353", strs("661"), nomore},
	{"/ 1 0", strs("+Inf"), nomore},
	// Equality
	{"put ?(== a a) ?(== [] []) ?(== [&] [&])",
		[]Value{
			Bool(true), Bool(false), Bool(false),
		}, nomore},
	{"kind-of bare 'str' [] [&] []{ }",
		strs("string", "string", "list", "map", "fn"), nomore},
	/*
		{"put ?(fail failed)", []Value{
			Error{&util.PosError{6, 17, errors.New("failed")}},
		}, nomore},
	*/
	{`put "l\norem" ipsum | into-lines`, strs(),
		more{wantBytesOut: []byte("l\norem\nipsum\n")}},
	{`echo "1\n233" | each put`, strs("1", "233"), nomore},
	{"put [a] [b c] | unpack", strs("a", "b", "c"), nomore},
	{`echo '{"k": "v", "a": [1, 2]}' '"foo"' | from-json`, []Value{
		NewMap(map[Value]Value{
			String("k"): String("v"),
			String("a"): NewList(strs("1", "2")...)}),
		String("foo"),
	}, nomore},
	{"base 16 42 233", strs("2a", "e9"), nomore},

	// eawk
	{`println "  ax  by cz  \n11\t22 33" | eawk { put $args[-1] }`,
		strs("cz", "33"), nomore},
	// Some builtins also take input from argument.
	{`each { put $0 } [x y z]`, strs("x", "y", "z"), nomore},
	{`count [1 2 3]`, strs("3"), nomore},
}

func strs(ss ...string) []Value {
	vs := make([]Value, len(ss))
	for i, s := range ss {
		vs[i] = String(s)
	}
	return vs
}

func bools(bs ...bool) []Value {
	vs := make([]Value, len(bs))
	for i, b := range bs {
		vs[i] = Bool(b)
	}
	return vs
}

func mustParse(t *testing.T, name, text string) *parse.Chunk {
	n, err := parse.Parse(text)
	if err != nil {
		t.Fatalf("Parser(%q) error: %s", text, err)
	}
	return n
}

func evalAndCollect(t *testing.T, texts []string, chsize int) ([]Value, []byte, error) {
	name := "<eval test>"
	ev := NewEvaler(nil)

	// Collect byte output
	outBytes := []byte{}
	pr, pw, _ := os.Pipe()
	bytesDone := make(chan struct{})
	go func() {
		for {
			var buf [64]byte
			nr, err := pr.Read(buf[:])
			outBytes = append(outBytes, buf[:nr]...)
			if err != nil {
				break
			}
		}
		close(bytesDone)
	}()

	// Channel output
	outs := []Value{}

	// Exit. Only the exit of the last text is saved.
	var ex error

	for _, text := range texts {
		n := mustParse(t, name, text)

		outCh := make(chan Value, chsize)
		outDone := make(chan struct{})
		go func() {
			for v := range outCh {
				outs = append(outs, v)
			}
			close(outDone)
		}()

		ports := []*Port{
			{File: os.Stdin},
			{File: pw, Chan: outCh},
			{File: os.Stderr},
		}

		ex = ev.Eval(name, text, n, ports)
		close(outCh)
		<-outDone
	}

	pw.Close()
	<-bytesDone
	return outs, outBytes, ex
}

func TestEval(t *testing.T) {
	for _, tt := range evalTests {
		// fmt.Printf("eval %q\n", tt.text)

		out, bytesOut, err := evalAndCollect(t, []string{tt.text}, len(tt.wantOut))

		good := true
		errorf := func(format string, args ...interface{}) {
			if good {
				good = false
				t.Errorf("eval(%q) fails:", tt.text)
			}
			t.Errorf(format, args...)
		}

		if string(bytesOut) != string(tt.wantBytesOut) {
			errorf("got bytesOut=%q, want %q", bytesOut, tt.wantBytesOut)
		}
		// Check error. We accept errAny as a "wildcard" for all non-nil errors.
		if !(tt.wantError == errAny && err != nil) && !reflect.DeepEqual(tt.wantError, err) {
			errorf("got err=%v, want %v", err, tt.wantError)
		}
		if !reflect.DeepEqual(tt.wantOut, out) {
			errorf("got out=%v, want %v", out, tt.wantOut)
		}
		if !good {
			t.Errorf("--------------")
		}
	}
}

func TestMultipleEval(t *testing.T) {
	texts := []string{"x=hello", "put $x"}
	outs, _, err := evalAndCollect(t, texts, 1)
	wanted := strs("hello")
	if err != nil {
		t.Errorf("eval %s => %v, want nil", texts, err)
	}
	if !reflect.DeepEqual(outs, wanted) {
		t.Errorf("eval %s outputs %v, want %v", texts, outs, wanted)
	}
}
