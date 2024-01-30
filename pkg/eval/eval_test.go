package eval_test

import (
	"sync"
	"testing"

	. "src.elv.sh/pkg/eval"

	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/parse"
)

func TestEval_AlternativeGlobal(t *testing.T) {
	ev := NewEvaler()
	g := BuildNs().AddVar("a", vars.NewReadOnly("")).Ns()
	err := ev.Eval(parse.Source{Name: "[test]", Code: "nop $a"}, EvalCfg{Global: g})
	if err != nil {
		t.Errorf("got error %v, want nil", err)
	}
	// Regression test for #1223
	if ev.Global().HasKeyString("a") {
		t.Errorf("$a from alternative global leaked into Evaler global")
	}
}

func TestEval_Concurrent(t *testing.T) {
	ev := NewEvaler()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		ev.Eval(parse.Source{Name: "[test]", Code: "var a"}, EvalCfg{})
		wg.Done()
	}()
	go func() {
		ev.Eval(parse.Source{Name: "[test]", Code: "var b"}, EvalCfg{})
		wg.Done()
	}()
	wg.Wait()
	g := ev.Global()
	if !g.HasKeyString("a") {
		t.Errorf("variable $a not created")
	}
	if !g.HasKeyString("b") {
		t.Errorf("variable $b not created")
	}
}

type fooOpts struct{ Opt string }

func (*fooOpts) SetDefaultOptions() {}

func TestCall(t *testing.T) {
	ev := NewEvaler()
	var gotOpt, gotArg string
	fn := NewGoFn("foo", func(fm *Frame, opts fooOpts, arg string) {
		gotOpt = opts.Opt
		gotArg = arg
	})

	passedArg := "arg value"
	passedOpt := "opt value"
	ev.Call(fn,
		CallCfg{
			Args: []any{passedArg},
			Opts: map[string]any{"opt": passedOpt},
			From: "[TestCall]"},
		EvalCfg{})

	if gotArg != passedArg {
		t.Errorf("got arg %q, want %q", gotArg, passedArg)
	}
	if gotOpt != passedOpt {
		t.Errorf("got opt %q, want %q", gotOpt, passedOpt)
	}
}

var checkTests = []struct {
	name           string
	code           string
	wantParseErr   bool
	wantCompileErr bool
}{
	{name: "no error", code: "put $nil"},
	{name: "parse error only", code: "put [",
		wantParseErr: true},
	{name: "compile error only", code: "put $x",
		wantCompileErr: true},
	{name: "both parse and compile error", code: "put [$x",
		wantParseErr: true, wantCompileErr: true},
}

func TestCheck(t *testing.T) {
	ev := NewEvaler()
	for _, test := range checkTests {
		t.Run(test.name, func(t *testing.T) {
			parseErr, _, compileErr := ev.Check(parse.Source{Name: "[test]", Code: test.code}, nil)
			if (parseErr != nil) != test.wantParseErr {
				t.Errorf("got parse error %v, when wantParseErr = %v",
					parseErr, test.wantParseErr)
			}
			if (compileErr != nil) != test.wantCompileErr {
				t.Errorf("got compile error %v, when wantCompileErr = %v",
					compileErr, test.wantCompileErr)
			}
		})
	}
}
