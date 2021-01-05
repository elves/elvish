package eval_test

import (
	"bytes"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/elves/elvish/pkg/eval"
	. "github.com/elves/elvish/pkg/eval"

	. "github.com/elves/elvish/pkg/eval/evaltest"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/prog"
	"github.com/elves/elvish/pkg/testutil"
)

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
	TestWithSetup(t,
		func(ev *Evaler) { ev.SetArgs([]string{"foo", "bar"}) },
		That("put $args").Puts(vals.MakeList("foo", "bar")))
}

func TestEvalTimeDeprecate(t *testing.T) {
	restore := prog.SetShowDeprecations(true)
	defer restore()
	_, cleanup := testutil.InTestDir()
	defer cleanup()

	TestWithSetup(t, func(ev *Evaler) {
		ev.AddGlobal(NsBuilder{}.AddGoFn("", "dep", func(fm *Frame) {
			fm.Deprecate("deprecated", nil)
		}).Ns())
	},
		That("dep").PrintsStderrWith("deprecated"),
		// Deprecation message is only shown once.
		That("dep 2> tmp.txt; dep").DoesNothing(),
	)
}

func TestCompileTimeDeprecation(t *testing.T) {
	restore := prog.SetShowDeprecations(true)
	defer restore()

	ev := NewEvaler()
	errOutput := new(bytes.Buffer)

	parseErr, compileErr := ev.Check(parse.Source{Code: "ord a"}, errOutput)
	if parseErr != nil {
		t.Errorf("got parse err %v", parseErr)
	}
	if compileErr != nil {
		t.Errorf("got compile err %v", compileErr)
	}

	warning := errOutput.String()
	wantWarning := `the "ord" command is deprecated`
	if !strings.Contains(warning, wantWarning) {
		t.Errorf("got warning %q, want warning containing %q", warning, wantWarning)
	}
}

func TestMultipleEval(t *testing.T) {
	texts := []string{"x=hello", "put $x"}
	r := EvalAndCollect(t, NewEvaler(), texts)
	wantOuts := []interface{}{"hello"}
	if r.Exception != nil {
		t.Errorf("eval %s => %v, want nil", texts, r.Exception)
	}
	if !reflect.DeepEqual(r.ValueOut, wantOuts) {
		t.Errorf("eval %s outputs %v, want %v", texts, r.ValueOut, wantOuts)
	}
}

func TestConcurrentEval(t *testing.T) {
	// Run this test with "go test -race".
	ev := NewEvaler()
	src := parse.Source{Name: "[test]", Code: ""}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		ev.Eval(src, EvalCfg{})
		wg.Done()
	}()
	go func() {
		ev.Eval(src, EvalCfg{})
		wg.Done()
	}()
	wg.Wait()
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
		eval.CallCfg{
			Args: []interface{}{passedArg},
			Opts: map[string]interface{}{"opt": passedOpt},
			From: "[TestCall]"},
		eval.EvalCfg{})

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
			parseErr, compileErr := ev.Check(parse.Source{Code: test.code}, nil)
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
