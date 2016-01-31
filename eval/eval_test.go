package eval

import (
	"errors"
	"os"
	"reflect"
	"strconv"
	"syscall"
	"testing"

	"github.com/elves/elvish/parse"
)

func TestNewEvaler(t *testing.T) {
	ev := NewEvaler(nil)
	pid := strconv.Itoa(syscall.Getpid())
	if ToString(ev.global["pid"].Get()) != pid {
		t.Errorf(`ev.global["pid"] = %v, want %v`, ev.global["pid"], pid)
	}
}

func strs(ss ...string) []Value {
	vs := make([]Value, len(ss))
	for i, s := range ss {
		vs[i] = String(s)
	}
	return vs
}

var anyerror = errors.New("")

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
	// Chunks
	// Empty chunk
	{"", []Value{}, nomore},
	// Outputs of pipelines in a chunk are concatenated
	{"put x; put y; put z", strs("x", "y", "z"), nomore},
	// A failed pipeline cause the whole chunk to fail
	{"put a; false; put b", strs("a"), more{wantError: errors.New("1")}},

	// Pipelines
	// Pure byte pipeline
	{`echo "Albert\nAllan\nAlbraham\nBerlin" | sed s/l/1/g | grep e`,
		[]Value{}, more{wantBytesOut: []byte("A1bert\nBer1in\n")}},
	// Pure channel pipeline
	{`put 233 42 19 | each [x]{+ $x 10}`, strs("243", "52", "29"), nomore},
	// TODO: Add a useful hybrid pipeline sample

	// Builtin functions
	// Arithmetics
	// TODO test more edge cases
	{"* 353 661", strs("233333"), nomore},
	{"+ 233100 233", strs("233333"), nomore},
	{"- 233333 233100", strs("233"), nomore},
	{"/ 233333 353", strs("661"), nomore},
	{"/ 1 0", strs("+Inf"), nomore},

	// String literal
	{`put 'such \"''literal'`, strs(`such \"'literal`), nomore},
	{`put "much \n\033[31;1m$cool\033[m"`,
		strs("much \n\033[31;1m$cool\033[m"), nomore},

	// Compounding
	{"put {fi,elvi}sh{1.0,1.1}",
		strs("fish1.0", "fish1.1", "elvish1.0", "elvish1.1"), nomore},

	// List, map and indexing
	{"println [a b c] [&key value] | from-lines",
		strs("[a b c] [&key value]"), nomore},
	{"put [a b c][2]", strs("c"), nomore},
	{"put [&key value][key]", strs("value"), nomore},

	// Output capture
	{"put (put lorem ipsum)", strs("lorem", "ipsum"), nomore},

	// Status capture
	{"put ?(true|false|false)",
		[]Value{newMultiError(OK, NewFailure("1"), NewFailure("1"))}, nomore},

	// Variable and compounding
	{"set x = 'SHELL'\nput 'WOW, SUCH '$x', MUCH COOL'\n",
		strs("WOW, SUCH SHELL, MUCH COOL"), nomore},

	// Closure
	// Basics
	{"[]{ }", strs(), nomore},
	{"[x]{put $x} foo", strs("foo"), nomore},
	// Variable capture
	{"set x = lorem; []{set x = ipsum}; put $x", strs("ipsum"), nomore},
	{"set x = lorem; []{ put $x; set x = ipsum }; put $x",
		strs("lorem", "ipsum"), nomore},
	// Shadowing
	{"set x = ipsum; []{ set local:x = lorem; put $x }; put $x",
		strs("lorem", "ipsum"), nomore},
	// Shadowing by argument
	{"set x = ipsum; [x]{ put $x; set x = BAD } lorem; put $x",
		strs("lorem", "ipsum"), nomore},
	// Closure captures new local variables every time
	{`fn f []{ set x = 0; put []{set x = (+ $x 1)} []{put $x} }
set inc1 put1 = (f); $put1; $inc1; $put1
set inc2 put2 = (f); $put2; $inc2; $put2`,
		strs("0", "1", "0", "1"), nomore},

	// fn
	{"fn f [x]{ put $x ipsum }; f lorem",
		strs("lorem", "ipsum"), nomore},
	/*
		// if
		{"if true; then put x", strs("x"), nomore},
		{"if true; false; then put x; else put y",
			strs("y"), nomore},
		{"if true; false; then put x; else if false; put y; else put z",
			strs("z"), nomore},
	*/

	// Namespaces
	// Pseudo-namespaces local: and up:
	{"set x = lorem; []{set local:x = ipsum; put $up:x $local:x}",
		strs("lorem", "ipsum"), nomore},
	{"set x = lorem; []{set up:x = ipsum; put $x}; put $x",
		strs("ipsum", "ipsum"), nomore},
	// Pseudo-namespace env:
	{"set env:foo = lorem; put $env:foo", strs("lorem"), nomore},
	{"del env:foo; put $env:foo", strs(""), nomore},
	// TODO: Test module namespace
}

func mustParse(t *testing.T, name, text string) *parse.Chunk {
	n, err := parse.Parse("<eval_test>", text)
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
	bytesExhausted := make(chan bool)
	go func() {
		for {
			var buf [64]byte
			nr, err := pr.Read(buf[:])
			outBytes = append(outBytes, buf[:nr]...)
			if err != nil {
				break
			}
		}
		bytesExhausted <- true
	}()

	// Channel output
	outs := []Value{}

	// Exit. Only the exit of the last text is saved.
	var ex error

	for _, text := range texts {
		n := mustParse(t, name, text)

		out := make(chan Value, chsize)
		exhausted := make(chan struct{})
		go func() {
			for v := range out {
				outs = append(outs, v)
			}
			exhausted <- struct{}{}
		}()

		ex = ev.evalWithOut(name, text, n, &port{ch: out, closeCh: true, f: pw})
		<-exhausted
	}
	pw.Close()
	<-bytesExhausted
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

		if tt.wantBytesOut != nil && !reflect.DeepEqual(tt.wantBytesOut, bytesOut) {
			errorf("got bytesOut=%q, want %q", bytesOut, tt.wantBytesOut)
		}
		if !(tt.wantError == anyerror && err != nil) && !reflect.DeepEqual(tt.wantError, err) {
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
	outs, _, err := evalAndCollect(t, []string{"set x = hello", "put $x"}, 1)
	wanted := strs("hello")
	if err != nil {
		t.Errorf("eval %q => %v, want nil", err)
	}
	if !reflect.DeepEqual(outs, wanted) {
		t.Errorf("eval %q outputs %v, want %v", outs, wanted)
	}
}
