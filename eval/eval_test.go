package eval

import (
	"os"
	"reflect"
	"strconv"
	"syscall"
	"testing"

	"github.com/elves/elvish/parse-ng"
)

func TestNewEvaler(t *testing.T) {
	ev := NewEvaler(nil, "")
	pid := strconv.Itoa(syscall.Getpid())
	if toString(ev.builtin["pid"].Get()) != pid {
		t.Errorf(`ev.builtin["pid"] = %v, want %v`, ev.builtin["pid"], pid)
	}
}

func strs(ss ...string) []Value {
	vs := make([]Value, len(ss))
	for i, s := range ss {
		vs[i] = str(s)
	}
	return vs
}

type more struct {
	wantBytesOut []byte
	wantExitus   []Value
	wantError    bool
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
	{"put a; false; put b", strs("a"), more{wantExitus: []Value{newFailure("1")}}},

	// Pipelines
	// Pure byte pipeline
	{`echo "Albert\nAllan\nAlbraham\nBerlin" | sed s/l/1/g | grep e`,
		[]Value{}, more{wantBytesOut: []byte("A1bert\nBer1in\n")}},
	// TODO: Add a useful channel pipeline sample
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
	{"println [a b c] [&key value] | feedchan",
		strs("[a b c][&key value]"), nomore},
	{"put [a b c][2]", strs("c"), nomore},
	{"put [&key value][key]", strs("value"), nomore},

	// Output capture
	{"put (put lorem ipsum)", strs("lorem", "ipsum"), nomore},

	// Status capture
	{"put ?(true|false|false)",
		[]Value{ok, newFailure("1"), newFailure("1")}, nomore},

	/*
		// Variable and compounding
		{"set x = `SHELL`\nput `WOW, SUCH `$x`, MUCH COOL`\n",
			strs("WOW, SUCH SHELL, MUCH COOL"), nomore},

		// Closure
		// Basics
		{"[]{ }", strs(), nomore},
		{"[x]{put $x} foo", strs("foo"), nomore},
		// Variable enclosure
		{"set x = lorem; []{ put $x; set x = ipsum }; put $x",
			strs("lorem", "ipsum"), nomore},
		// Shadowing
		{"set x = ipsum; []{ var x = lorem; put $x }; put $x",
			strs("lorem", "ipsum"), nomore},
		// Shadowing by argument
		{"set x = ipsum; [x]{ put $x; set x = BAD } lorem; put $x",
			strs("lorem", "ipsum"), nomore},

		// fn
		{"fn f [x]{ put $x ipsum }; f lorem",
			strs("lorem", "ipsum"), nomore},
		// if
		{"if true; then put x", strs("x"), nomore},
		{"if true; false; then put x; else put y",
			strs("y"), nomore},
		{"if true; false; then put x; else if false; put y; else put z",
			strs("z"), nomore},

		// Namespaces
		// Pseudo-namespaces local: and up:
		{"set true = lorem; []{set true = ipsum; put $up:true $local:true $builtin:true}",
			[]Value{str("lorem"), str("ipsum"), boolean(true)}, nomore},
		{"set x = lorem; []{set up:x = ipsum}; put x", strs("ipsum"), nomore},
		// Pseudo-namespace env:
		{"set env:foo = lorem; put $env:foo", strs("lorem"), nomore},
		{"del env:foo; put $env:foo", strs(""), nomore},
		// TODO: Test module namespace
	*/
}

func mustParse(t *testing.T, name, text string) *parse.Chunk {
	n, err := parse.Parse("<eval_test>", text)
	if err != nil {
		t.Fatalf("Parser(%q) error: %s", text, err)
	}
	return n
}

func evalAndCollect(t *testing.T, texts []string, chsize int) ([]Value, []byte, []Value, error) {
	name := "<eval test>"
	ev := NewEvaler(nil, ".")

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

	// Exitus. Only the exitus of the last text is saved.
	var ex []Value

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

		var err error
		ex, err = ev.evalWithOut(name, text, n, &port{ch: out, closeCh: true, f: pw})
		if err != nil {
			return outs, outBytes, ex, err
		}
		<-exhausted
	}
	pw.Close()
	<-bytesExhausted
	return outs, outBytes, ex, nil
}

func TestEval(t *testing.T) {
	for _, tt := range evalTests {
		out, bytesOut, ex, err := evalAndCollect(t, []string{tt.text}, len(tt.wantOut))

		if (tt.wantError) != (err != nil) ||
			(tt.wantBytesOut != nil && !reflect.DeepEqual(tt.wantBytesOut, bytesOut)) ||
			(tt.wantExitus != nil && !reflect.DeepEqual(tt.wantExitus, ex)) ||
			!reflect.DeepEqual(tt.wantOut, out) {

			wantError := "nil"
			if tt.wantError {
				wantError = "non-nil"
			}
			t.Errorf("eval %q out=%v, bytesOut=%q, exitus=%v, err=%v; want out=%v, bytesOut=%q, exitus=%v, err=%s",
				tt.text, out, bytesOut, ex, err, tt.wantOut, tt.wantBytesOut, tt.wantExitus, wantError)
		}
	}
}

/*
func TestMultipleEval(t *testing.T) {
	outs, err := evalAndCollect([]string{"var $x = `hello`", "put $x"}, 1)
	wanted := strs("hello")
	if err != nil {
		t.Errorf("eval %q => %v, want nil", err)
	}
	if !reflect.DeepEqual(outs, wanted) {
		t.Errorf("eval %q outputs %v, want %v", outs, wanted)
	}
}
*/
