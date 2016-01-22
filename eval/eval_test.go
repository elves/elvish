package eval

import (
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

func stringValues(ss ...string) []Value {
	vs := make([]Value, len(ss))
	for i, s := range ss {
		vs[i] = str(s)
	}
	return vs
}

var evalTests = []struct {
	text      string
	wanted    []Value
	wantError bool
}{
	// Empty chunk
	{"", []Value{}, false},

	// Trivial command
	{"put 233 lorem ipsum", stringValues("233", "lorem", "ipsum"), false},

	// Byte Pipeline
	{`echo "Albert\nAllan\nAlbraham\nBerlin" | sed s/l/1/g | grep e | feedchan`,
		stringValues("A1bert", "Ber1in"), false},

	// Arithmetics
	// TODO test more edge cases
	{"* 353 661", stringValues("233333"), false},
	{"+ 233100 233", stringValues("233333"), false},
	{"- 233333 233100", stringValues("233"), false},
	{"/ 233333 353", stringValues("661"), false},
	{"/ 1 0", stringValues("+Inf"), false},

	// String literal
	{`put 'such \"''literal'`, stringValues(`such \"'literal`), false},
	{`put "much \n\033[31;1m$cool\033[m"`,
		stringValues("much \n\033[31;1m$cool\033[m"), false},

	// Compounding list primaries
	{"put {fi,elvi}sh{1.0,1.1}",
		stringValues("fish1.0", "fish1.1", "elvish1.0", "elvish1.1"), false},

	// List, map and indexing
	{"println [a b c] [&key value] | feedchan",
		stringValues("[a b c][&key value]"), false},
	{"put [a b c][2]", stringValues("c"), false},
	{"put [&key value][key]", stringValues("value"), false},

	// Variable and compounding
	{"set x = `SHELL`\nput `WOW, SUCH `$x`, MUCH COOL`\n",
		stringValues("WOW, SUCH SHELL, MUCH COOL"), false},

	// Channel capture
	{"put (put lorem ipsum)", stringValues("lorem", "ipsum"), false},

	// Status capture
	{"put ?(true|false|false)",
		[]Value{ok, newFailure("1"), newFailure("1")}, false},

	// Closure evaluation
	{"[]{ }", stringValues(), false},
	{"[x]{put $x} foo", stringValues("foo"), false},

	// Variable enclosure
	{"set x = lorem; []{ put $x; set x = ipsum }; put $x",
		stringValues("lorem", "ipsum"), false},
	// Shadowing
	{"set x = ipsum; []{ var x = lorem; put $x }; put $x",
		stringValues("lorem", "ipsum"), false},
	// Shadowing by argument
	{"set x = ipsum; [x]{ put $x; set x = BAD } lorem; put $x",
		stringValues("lorem", "ipsum"), false},

	// fn
	{"fn f [x]{ put $x ipsum }; f lorem",
		stringValues("lorem", "ipsum"), false},
	// if
	{"if true; then put x", stringValues("x"), false},
	{"if true; false; then put x; else put y",
		stringValues("y"), false},
	{"if true; false; then put x; else if false; put y; else put z",
		stringValues("z"), false},

	// Namespaces
	// Pseudo-namespaces local: and up:
	{"set true = lorem; []{set true = ipsum; put $up:true $local:true $builtin:true}",
		[]Value{str("lorem"), str("ipsum"), boolean(true)}, false},
	{"set x = lorem; []{set up:x = ipsum}; put x", stringValues("ipsum"), false},
	// Pseudo-namespace env:
	{"set env:foo = lorem; put $env:foo", stringValues("lorem"), false},
	{"del env:foo; put $env:foo", stringValues(""), false},
}

func mustParse(t *testing.T, name, text string) *parse.Chunk {
	n, err := parse.Parse("<eval_test>", text)
	if err != nil {
		t.Fatalf("Parser(%q) error: %s", text, err)
	}
	return n
}

func evalAndCollect(t *testing.T, texts []string, chsize int) ([]Value, error) {
	name := "<eval test>"
	ev := NewEvaler(nil, ".")
	outs := []Value{}

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

		// XXX(xiaq): Ignore failures
		_, err := ev.evalWithChanOut(name, text, n, out)
		if err != nil {
			return outs, err
		}
		close(out)
		<-exhausted
	}
	return outs, nil
}

func TestEval(t *testing.T) {
	for _, tt := range evalTests {
		outs, err := evalAndCollect(t, []string{tt.text}, len(tt.wanted))

		if tt.wantError {
			// Test for error, ignore output
			if err == nil {
				t.Errorf("eval %q => <nil>, want non-nil", tt.text)
			}
		} else {
			// Test for output
			if err != nil {
				t.Errorf("eval %q => %v, want <nil>", tt.text, err)
			}
			if !reflect.DeepEqual(outs, tt.wanted) {
				t.Errorf("Evalling %q outputs %v, want %v", tt.text, outs, tt.wanted)
			}
		}
	}
}

/*
func TestMultipleEval(t *testing.T) {
	outs, err := evalAndCollect([]string{"var $x = `hello`", "put $x"}, 1)
	wanted := stringValues("hello")
	if err != nil {
		t.Errorf("eval %q => %v, want nil", err)
	}
	if !reflect.DeepEqual(outs, wanted) {
		t.Errorf("eval %q outputs %v, want %v", outs, wanted)
	}
}
*/
