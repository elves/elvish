package eval

import (
	"reflect"
	"strconv"
	"syscall"
	"testing"

	"github.com/elves/elvish/parse"
)

func TestNewEvaluator(t *testing.T) {
	ev := NewEvaluator()
	pid := strconv.Itoa(syscall.Getpid())
	if (*ev.builtin["pid"].valuePtr).String() != pid {
		t.Errorf(`ev.builtin["pid"] = %v, want %v`, ev.builtin["pid"], pid)
	}
}

func mustParse(name, text string) *parse.ChunkNode {
	n, e := parse.Parse(name, text)
	if e != nil {
		panic("parser error")
	}
	return n
}

func stringValues(ss ...string) []Value {
	vs := make([]Value, len(ss))
	for i, s := range ss {
		vs[i] = NewString(s)
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
	{"put `such \\\"``literal`", stringValues("such \\\"`literal"), false},
	{`put "much \n\033[31;1m$cool\033[m"`,
		stringValues("much \n\033[31;1m$cool\033[m"), false},

	// Compounding list primaries
	{"put {fi elvi}sh{1 2}",
		stringValues("fish1", "fish2", "elvish1", "elvish2"), false},

	// Table and subscript
	{"println [a b c &key value] | feedchan",
		stringValues("[a b c &key value]"), false},
	{"put [a b c &key value][2]", stringValues("c"), false},
	{"put [a b c &key value][key]", stringValues("value"), false},

	// Variable and compounding
	{"var $x string = `SHELL`\nput `WOW, SUCH `$x`, MUCH COOL`\n",
		stringValues("WOW, SUCH SHELL, MUCH COOL"), false},

	// var and set
	{"var $x $y string = SUCH VAR; put $x $y",
		stringValues("SUCH", "VAR"), false},
	{"var $x $y string; set $x $y = SUCH SET; put $x $y",
		stringValues("SUCH", "SET"), false},
	{"var $x", stringValues(), false},
	{"var $x string $y", stringValues(), false},
	{"var $x table; set $x = [lorem ipsum]; put $x[1]",
		stringValues("ipsum"), false},

	// Channel capture
	{"put (put lorem ipsum)", stringValues("lorem", "ipsum"), false},

	// Status capture
	{"put ?(true|false|false)",
		[]Value{success, newFailure("1"), newFailure("1")}, false},

	// Closure evaluation
	{"{ }", stringValues(), false},
	{"{|$x| put $x} foo", stringValues("foo"), false},

	// Variable enclosure
	{"var $x = lorem; { put $x; set $x = ipsum }; put $x",
		stringValues("lorem", "ipsum"), false},
	// Shadowing
	{"var $x = ipsum; { var $x = lorem; put $x }; put $x",
		stringValues("lorem", "ipsum"), false},
	// Shadowing by argument
	{"var $x = ipsum; { |$x| put $x; set $x = BAD } lorem; put $x",
		stringValues("lorem", "ipsum"), false},

	// fn
	{"fn f $x { put $x ipsum }; f lorem",
		stringValues("lorem", "ipsum"), false},

	// Namespace
	{"var $true = lorem; { var $true = ipsum; put $captured:true $local:true $builtin:true }",
		[]Value{String("lorem"), String("ipsum"), Bool(true)}, false},
	{"var $x = lorem; { set $captured:x = ipsum }; put $x",
		stringValues("ipsum"), false},
}

func TestEval(t *testing.T) {
	name := "<eval test>"
	for _, tt := range evalTests {
		n := mustParse(name, tt.text)

		ev := NewEvaluator()
		out := make(chan Value, len(tt.wanted))
		outs := []Value{}
		exhausted := make(chan struct{})
		go func() {
			for v := range out {
				outs = append(outs, v)
			}
			exhausted <- struct{}{}
		}()

		ev.ports[1].ch = out

		e := ev.Eval(name, tt.text, n)
		close(out)
		<-exhausted
		if tt.wantError {
			// Test for error, ignore output
			if e == nil {
				t.Errorf("ev.Eval(*, %q, *) => <nil>, want non-nil", tt.text)
			}
		} else {
			// Test for output
			if e != nil {
				t.Errorf("ev.Eval(*, %q, *) => %v, want <nil>", tt.text, e)
			}
			if !reflect.DeepEqual(outs, tt.wanted) {
				t.Errorf("Evalling %q outputs %v, want %v", tt.text, outs, tt.wanted)
			}
		}
	}
}
