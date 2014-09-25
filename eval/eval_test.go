package eval

import (
	"reflect"
	"strconv"
	"syscall"
	"testing"

	"github.com/xiaq/elvish/parse"
)

func TestNewEvaluator(t *testing.T) {
	ev := NewEvaluator()
	pid := strconv.Itoa(syscall.Getpid())
	if (*ev.scope["pid"]).String() != pid {
		t.Errorf(`ev.scope["pid"] = %v, want %v`, ev.scope["pid"], pid)
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
	{"var $x string `SHELL`\nput `WOW, SUCH `$x`, MUCH COOL`\n",
		stringValues("WOW, SUCH SHELL, MUCH COOL"), false},

	// var and set
	{"var [$x $y string] [SUCH VAR]; put $x $y",
		stringValues("SUCH", "VAR"), false},
	{"var [$x $y string]; set [$x $y] [SUCH SET]; put $x $y",
		stringValues("SUCH", "SET"), false},
	{"var $x", nil, true},
	{"var [$x string $y]", nil, true},

	// Status capture
	{"put ?(true|false|false)",
		stringValues("", "exited 1", "exited 1"), false},
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
