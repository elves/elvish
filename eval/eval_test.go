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
	text   string
	wanted []Value
}{
	{"", []Value{}},
	{"put 233", stringValues("233")},
	{"put {fi elvi}sh{1 2}", stringValues("fish1", "fish2", "elvish1", "elvish2")},
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

		ev.ports[1] = &port{ch: out}

		e := ev.Eval(name, tt.text, n)
		close(out)
		if e != nil {
			t.Errorf("ev.Eval(*, %q, *) => %v, want <nil>", tt.text, e)
		}
		<-exhausted
		if !reflect.DeepEqual(outs, tt.wanted) {
			t.Errorf("Evalling %q outputs %v, want %v", tt.text, outs, tt.wanted)
		}
	}
}
