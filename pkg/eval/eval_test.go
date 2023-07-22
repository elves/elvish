package eval_test

import (
	"strconv"
	"sync"
	"syscall"
	"testing"

	. "src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/prog"

	. "src.elv.sh/pkg/eval/evaltest"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/testutil"
)

func TestPid(t *testing.T) {
	pid := strconv.Itoa(syscall.Getpid())
	Test(t, That("put $pid").Puts(pid))
}

func TestNumBgJobs(t *testing.T) {
	Test(t,
		That("put $num-bg-jobs").Puts("0"),
		// TODO(xiaq): Test cases where $num-bg-jobs > 0. This cannot be done
		// with { put $num-bg-jobs }& because the output channel may have
		// already been closed when the closure is run.
	)
}

func TestArgs(t *testing.T) {
	Test(t,
		That("put $args").Puts(vals.EmptyList))
	TestWithEvalerSetup(t,
		func(ev *Evaler) { ev.Args = vals.MakeList("foo", "bar") },
		That("put $args").Puts(vals.MakeList("foo", "bar")))
}

func TestEvalTimeDeprecate(t *testing.T) {
	testutil.Set(t, &prog.DeprecationLevel, 42)
	testutil.InTempDir(t)

	TestWithEvalerSetup(t, func(ev *Evaler) {
		ev.ExtendGlobal(BuildNs().AddGoFn("dep", func(fm *Frame) {
			fm.Deprecate("deprecated", nil, 42)
		}))
	},
		That("dep").PrintsStderrWith("deprecated"),
		// Deprecation message from the same location is only shown once.
		That("fn f { dep }", "f 2> tmp.txt; f").DoesNothing(),
	)
}

func TestMultipleEval(t *testing.T) {
	Test(t,
		That("var x = hello").Then("put $x").Puts("hello"),

		// Shadowing with fn. Regression test for #1213.
		That("fn f { put old }").Then("fn f { put new }").Then("f").
			Puts("new"),
		// Variable deletion. Regression test for #1213.
		That("var x = foo").Then("del x").Then("put $x").DoesNotCompile("variable $x not found"),
	)
}

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
